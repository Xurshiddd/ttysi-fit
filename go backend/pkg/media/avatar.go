// Package media — tashqi rasmlarni (HEMIS avatarlari) yuklab olib, lokal
// diskka saqlash va ularga ommaviy URL berish uchun.
package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// avatarsSubdir — Dir ichidagi avatar fayllar papkasi.
const avatarsSubdir = "avatars"

// extByContentType — image content-type → fayl kengaytmasi.
var extByContentType = map[string]string{
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// Downloader — rasmlarni yuklab olib lokal diskka saqlaydi.
// Goroutine-safe: bir nechta goroutine bir vaqtda Save chaqirishi mumkin
// (har bir yozuv vaqtinchalik faylga, keyin atomik rename qilinadi).
type Downloader struct {
	http        *http.Client
	dir         string // avatar fayllar papkasi (Dir/avatars)
	routePrefix string // masalan "/static"
	publicBase  string // masalan "http://localhost:8080" yoki "" (nisbiy)
	maxBytes    int64
	// allowedHosts — SSRF himoyasi: faqat shu hostlardan yuklab olinadi
	// (CLAUDE.md §17.3 #7). Bo'sh bo'lsa allowlist o'chiriladi, lekin
	// private/loopback IP bloklash baribir ishlaydi (ssrfDialControl).
	allowedHosts map[string]struct{}
}

// NewDownloader — yuklab oluvchini yaratadi va saqlash papkasini tayyorlaydi.
//
// allowedHosts — ruxsat etilgan hostlar (masalan HEMIS domenlari). SSRF
// himoyasi ikki qatlamda: (1) URL host allowlist'da bo'lishi shart,
// (2) DNS resolve'dan KEYIN ulanish bosqichida private/loopback/link-local
// IP'lar rad etiladi (DNS rebinding'ga ham himoya). Redirect'larning har
// bir bosqichi ham xuddi shu tekshiruvdan o'tadi.
func NewDownloader(dir, routePrefix, publicBase string, maxBytes int64, timeout time.Duration, allowedHosts []string) (*Downloader, error) {
	if maxBytes <= 0 {
		maxBytes = 5 << 20
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	avatarsDir := filepath.Join(dir, avatarsSubdir)
	if err := os.MkdirAll(avatarsDir, 0o755); err != nil {
		return nil, fmt.Errorf("media: papka yaratilmadi: %w", err)
	}

	d := &Downloader{
		dir:          avatarsDir,
		routePrefix:  strings.TrimRight(routePrefix, "/"),
		publicBase:   strings.TrimRight(publicBase, "/"),
		maxBytes:     maxBytes,
		allowedHosts: make(map[string]struct{}, len(allowedHosts)),
	}
	for _, h := range allowedHosts {
		if h = strings.ToLower(strings.TrimSpace(h)); h != "" {
			d.allowedHosts[h] = struct{}{}
		}
	}

	d.http = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
				Control: ssrfDialControl, // resolve qilingan IP tekshiruvi
			}).DialContext,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("media: juda ko'p redirect")
			}
			return d.validateURL(req.URL.String()) // har bir redirect ham allowlist'dan o'tadi
		},
	}
	return d, nil
}

// validateURL — SSRF himoyasi: sxema http/https va host allowlist tekshiruvi.
func (d *Downloader) validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("media: URL noto'g'ri: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("media: taqiqlangan sxema (%s)", u.Scheme)
	}
	if len(d.allowedHosts) > 0 {
		if _, ok := d.allowedHosts[strings.ToLower(u.Hostname())]; !ok {
			return fmt.Errorf("media: ruxsat etilmagan host (%s)", u.Hostname())
		}
	}
	return nil
}

// ssrfDialControl — TCP ulanishdan oldin (DNS resolve'dan keyin) chaqiriladi:
// private (10/172.16/192.168, fc00::/7), loopback (127.x, ::1), link-local
// (169.254.x — cloud metadata 169.254.169.254 ham shu yerda) va boshqa
// global bo'lmagan IP'larga ulanish rad etiladi.
func ssrfDialControl(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("media: manzil noto'g'ri: %w", err)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("media: IP aniqlanmadi (%s)", host)
	}
	if !ip.IsGlobalUnicast() || ip.IsPrivate() {
		return fmt.Errorf("media: taqiqlangan IP (%s)", ip)
	}
	return nil
}

// Save — srcURL dagi rasmni yuklab olib, avatars/<name>.<ext> sifatida saqlaydi
// va uning ommaviy URL'ini qaytaradi.
//
// Idempotent: <name> uchun fayl allaqachon mavjud bo'lsa, qayta yuklamaydi —
// mavjud faylning URL'ini qaytaradi. Shu sababli takroriy sync tez ishlaydi.
//
// srcURL bo'sh bo'lsa ("", nil) qaytadi — chaqiruvchi default avatarni hal qiladi.
func (d *Downloader) Save(ctx context.Context, srcURL, name string) (string, error) {
	if srcURL == "" {
		return "", nil
	}
	name = sanitize(name)
	if name == "" {
		return "", errors.New("media: bo'sh fayl nomi")
	}

	// SSRF himoyasi — yuklab olishdan oldin URL tekshiruvi (§17.3 #7).
	if err := d.validateURL(srcURL); err != nil {
		return "", err
	}

	// 1. Mavjud faylni qidirish (idempotent).
	if fname := d.findExisting(name); fname != "" {
		return d.publicURL(fname), nil
	}

	// 2. So'rov.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srcURL, nil)
	if err != nil {
		return "", fmt.Errorf("media: so'rov yaratilmadi: %w", err)
	}
	req.Header.Set("Accept", "image/*")
	// Ba'zi serverlar User-Agent'siz so'rovlarni bloklaydi — brauzerga o'xshatamiz.
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ttysi-fit/1.0)")

	res, err := d.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("media: yuklab olish xatosi: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("media: status %d (%s)", res.StatusCode, srcURL)
	}

	// 3. Content-type validatsiyasi (CLAUDE.md §3.2 — fayl turi tekshiruvi).
	ext, ok := extFromContentType(res.Header.Get("Content-Type"))
	if !ok {
		// Header ishonchsiz bo'lsa — URL kengaytmasiga tushamiz.
		ext = extFromURL(srcURL)
	}
	if ext == "" {
		return "", fmt.Errorf("media: rasm turi aniqlanmadi (%s)", srcURL)
	}

	// 4. Hajm cheklovi bilan o'qish (maxBytes+1 — oshib ketganini aniqlash uchun).
	limited := io.LimitReader(res.Body, d.maxBytes+1)

	// 5. Atomik yozuv: avval *.tmp, keyin rename.
	finalName := name + ext
	finalPath := filepath.Join(d.dir, finalName)
	tmp, err := os.CreateTemp(d.dir, name+"-*.tmp")
	if err != nil {
		return "", fmt.Errorf("media: temp fayl: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		tmp.Close()
		_ = os.Remove(tmpPath) // muvaffaqiyatdan keyin rename qilinadi, bu no-op bo'ladi
	}()

	written, err := io.Copy(tmp, limited)
	if err != nil {
		return "", fmt.Errorf("media: yozish xatosi: %w", err)
	}
	if written > d.maxBytes {
		return "", fmt.Errorf("media: rasm juda katta (>%d bayt)", d.maxBytes)
	}
	if written == 0 {
		return "", errors.New("media: bo'sh rasm")
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("media: yopish xatosi: %w", err)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return "", fmt.Errorf("media: rename xatosi: %w", err)
	}

	return d.publicURL(finalName), nil
}

// findExisting — <name>.<har qanday kengaytma> faylini qidiradi.
func (d *Downloader) findExisting(name string) string {
	for _, ext := range []string{".jpg", ".png", ".webp", ".gif"} {
		fname := name + ext
		if _, err := os.Stat(filepath.Join(d.dir, fname)); err == nil {
			return fname
		}
	}
	return ""
}

// publicURL — saqlangan fayl uchun ommaviy URL quradi.
func (d *Downloader) publicURL(fname string) string {
	path := d.routePrefix + "/" + avatarsSubdir + "/" + fname
	if d.publicBase != "" {
		return d.publicBase + path
	}
	return path
}

// extFromContentType — "image/jpeg; charset=..." dan kengaytma ajratadi.
func extFromContentType(ct string) (string, bool) {
	if ct == "" {
		return "", false
	}
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		mt = strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
	}
	ext, ok := extByContentType[strings.ToLower(mt)]
	return ext, ok
}

// extFromURL — URL yo'lidagi kengaytmani (.jpg/.png/...) qaytaradi.
func extFromURL(rawURL string) string {
	// Query/fragmentni tashlab yuboramiz.
	if i := strings.IndexAny(rawURL, "?#"); i >= 0 {
		rawURL = rawURL[:i]
	}
	ext := strings.ToLower(filepath.Ext(rawURL))
	switch ext {
	case ".jpg", ".jpeg":
		return ".jpg"
	case ".png", ".webp", ".gif":
		return ext
	}
	return ""
}

// sanitize — fayl nomini xavfsiz belgilarga cheklaydi (path traversal oldini olish).
func sanitize(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		}
	}
	return b.String()
}
