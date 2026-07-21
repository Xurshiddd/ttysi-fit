// Package certificate — yutuq sertifikatini PDF sifatida chizadi.
//
// Sertifikat DISKKA SAQLANMAYDI: har so'ralganda qaytadan chiziladi. Sabab —
// shablon o'zgarganda avval berilgan sertifikatlar ham darrov yangilanadi va
// fayl boshqaruvi (eskirgan fayllarni tozalash) umuman kerak bo'lmaydi.
// Chizish ~10 ms, sertifikat esa kamdan-kam yuklab olinadi.
package certificate

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

// Shriftlar binarga singdirilgan: serverda alohida fayl talab qilinmaydi.
//
// NEGA DejaVu: Go'ning o'z shrifti (golang.org/x/image/font/gofont) o'zbek
// kirillchasidagi Қ, Ғ, Ҳ va lotin ʻ (U+02BB) harflarini QOPLAMAYDI — HEMIS'dan
// kelgan ismlar bo'sh joy bo'lib chiqardi. DejaVu hammasini qoplaydi.
// Litsenziya: fonts/LICENSE-DejaVu.txt (Bitstream Vera — tarqatish mumkin).
//
//go:embed fonts/DejaVuSans.ttf
var fontRegular []byte

//go:embed fonts/DejaVuSans-Bold.ttf
var fontBold []byte

// Institut emblemasi (ttysi.uz rasmiy logotipi, SVG'dan PNG ga o'girilgan).
// Alfa kanali bor — sertifikat foniga to'g'ri tushadi.
//
//go:embed assets/ttysi-logo.png
var logoPNG []byte

// ─────────────────────── Shablon sozlamalari ───────────────────────
// Ko'rinishni o'zgartirish uchun asosan shu blokni tahrirlash kifoya.

const (
	// Sahifa: A4 albom (landscape).
	pageOrientation = "L"
	pageSize        = "A4"
	pageW           = 297.0 // mm
	pageH           = 210.0 // mm
	margin          = 15.0

	fontName = "DejaVu"

	// Emblema o'lchamlari (mm).
	logoW = 16.0 // sarlavhadagi emblema kengligi
	logoH = 18.9 // 479×566 nisbatini saqlaydi
)

// SUV BELGISI YO'Q — ataylab.
//
// Emblemani fon suv belgisi sifatida sinab ko'rildi va olib tashlandi: qalqon
// shakli tepasi to'g'ri burchakli, shuning uchun xira holatda ham matnni kesib
// o'tuvchi to'rtburchak chekka hosil qiladi va dizayn xatosidek ko'rinadi.
// Sarlavhadagi emblema o'zi yetarli.

// Ranglar (CLAUDE.md §8.1 dizayn tizimi).
var (
	colorPrimary = rgb{30, 58, 95}    // #1E3A5F chuqur ko'k
	colorAccent  = rgb{0, 200, 150}   // #00C896 yashil
	colorText    = rgb{33, 37, 41}    // asosiy matn
	colorMuted   = rgb{120, 130, 145} // ikkilamchi matn
)

type rgb struct{ R, G, B int }

// Matnlar — tarjima/qayta yozish uchun bir joyda.
const (
	textCertificate = "SERTIFIKAT"
	textAwardedTo   = "Ushbu sertifikat bilan taqdirlanadi"
	textFor         = "quyidagi yutuq uchun"
	textSignature   = "Institut rahbariyati"
	textNumberLabel = "Sertifikat"
	textDateLabel   = "Berilgan sana"
	// textSince — institut shiori (rasmiy saytdan: "1932-yildan beri").
	textSince = "1932-yildan beri"

	// NAMUNA muhri yozuvlari — hujjat haqiqiy emasligi ko'rinib tursin.
	textSampleStamp     = "NAMUNA"
	textSampleStampNote = "haqiqiy muhr emas"
)

// DefaultOrganization — tashkilot nomi ko'rsatilmaganda ishlatiladi.
const DefaultOrganization = "Toshkent to'qimachilik va yengil sanoat instituti"

// ─────────────────────────── Ma'lumot ───────────────────────────

// Data — sertifikatga chiqadigan ma'lumot.
type Data struct {
	Organization string // bo'sh bo'lsa DefaultOrganization

	FullName    string // kimga berilyapti
	Title       string // yutuq nomi
	Description string // yutuq izohi (ixtiyoriy)

	// ValueLabel — yutuq o'lchovi, masalan "100 000 qadam" (ixtiyoriy).
	ValueLabel string

	Number    string    // sertifikat raqami (odatda qisqartirilgan UUID)
	AwardedAt time.Time // berilgan sana
	Note      string    // qo'lda berilganda admin izohi (ixtiyoriy)

	// Signing — muhr, imzo va imzolovchi. Nol qiymatda hozirgidek faqat
	// bo'sh imzo chizig'i chiziladi (sertifikat baribir beriladi).
	Signing Signing
}

// Render — sertifikatni PDF bayt massivi sifatida qaytaradi.
func Render(d Data) ([]byte, error) {
	if d.Organization == "" {
		d.Organization = DefaultOrganization
	}
	if d.AwardedAt.IsZero() {
		d.AwardedAt = time.Now()
	}

	pdf := fpdf.New(pageOrientation, "mm", pageSize, "")
	pdf.SetAutoPageBreak(false, 0) // sertifikat — qat'iy bitta sahifa
	pdf.SetTitle(fmt.Sprintf("%s — %s", textCertificate, d.FullName), true)

	// UTF-8 shriftlar (fpdf'ning o'rnatilgan shriftlari faqat Latin-1).
	pdf.AddUTF8FontFromBytes(fontName, "", fontRegular)
	pdf.AddUTF8FontFromBytes(fontName, "B", fontBold)

	// Emblema bir marta ro'yxatdan o'tkaziladi, ikki joyda (sarlavha va suv
	// belgisi) qayta ishlatiladi — PDF ichida rasm bir nusxada saqlanadi.
	pdf.RegisterImageOptionsReader(
		logoName,
		fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false},
		bytes.NewReader(logoPNG),
	)

	// Muhr va imzo — faqat berilgan bo'lsa ro'yxatdan o'tkaziladi.
	if len(d.Signing.StampPNG) > 0 {
		pdf.RegisterImageOptionsReader(stampName,
			fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false},
			bytes.NewReader(d.Signing.StampPNG))
	}
	if len(d.Signing.SignaturePNG) > 0 {
		pdf.RegisterImageOptionsReader(signatureName,
			fpdf.ImageOptions{ImageType: "PNG", ReadDpi: false},
			bytes.NewReader(d.Signing.SignaturePNG))
	}

	pdf.AddPage()
	drawFrame(pdf)
	y := drawHeader(pdf, d)
	y = drawBody(pdf, d, y)
	drawFooter(pdf, d, y)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("certificate: PDF chizish: %w", err)
	}
	return buf.Bytes(), nil
}

// fpdf ichidagi rasm identifikatorlari.
const (
	logoName      = "ttysi-logo"
	stampName     = "cert-stamp"
	signatureName = "cert-signature"
)

// Imzo bloki o'lchamlari (mm).
//
// O'lchamlar ATAYLAB kichik: sertifikat matni (tavsif) sahifaning pastigacha
// tushishi mumkin, muhr esa uning ustiga chiqib ketmasligi kerak. Muhr
// markazi imzo chizig'idan PASTDA turadi — shunda u faqat chiziq va bo'sh
// joyni qoplaydi, matnni emas.
const (
	stampR     = 12.0 // muhr radiusi
	signatureW = 36.0 // imzo skani kengligi
	signatureH = 11.0 // imzo skani balandligi
)

// drawFrame — ikki qavatli ramka (tashqi qalin ko'k, ichki ingichka yashil).
func drawFrame(pdf *fpdf.Fpdf) {
	setDraw(pdf, colorPrimary)
	pdf.SetLineWidth(1.6)
	pdf.Rect(margin, margin, pageW-2*margin, pageH-2*margin, "D")

	setDraw(pdf, colorAccent)
	pdf.SetLineWidth(0.4)
	in := margin + 4
	pdf.Rect(in, in, pageW-2*in, pageH-2*in, "D")
}

// drawHeader — tashkilot nomi, sarlavha va ajratuvchi chiziq.
// Qaytaradi: keyingi blok boshlanadigan Y.
func drawHeader(pdf *fpdf.Fpdf, d Data) float64 {
	y := margin + 8

	// Institut emblemasi — markazda, tashkilot nomi tepasida.
	pdf.ImageOptions(logoName, (pageW-logoW)/2, y, logoW, logoH,
		false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	y += logoH + 3

	setText(pdf, colorMuted)
	pdf.SetFont(fontName, "", 11)
	centered(pdf, y, d.Organization)

	y += 6
	pdf.SetFont(fontName, "", 8)
	centered(pdf, y, textSince)

	y += 12
	setText(pdf, colorPrimary)
	pdf.SetFont(fontName, "B", 38)
	centered(pdf, y, textCertificate)

	// Sarlavha ostidagi qisqa aksent chiziq.
	y += 14
	setDraw(pdf, colorAccent)
	pdf.SetLineWidth(1.2)
	pdf.Line(pageW/2-25, y, pageW/2+25, y)

	return y + 12
}

// drawBody — kim, nima uchun, qancha.
func drawBody(pdf *fpdf.Fpdf, d Data, y float64) float64 {
	setText(pdf, colorMuted)
	pdf.SetFont(fontName, "", 12)
	centered(pdf, y, textAwardedTo)

	// Ism — sertifikatning asosiy elementi.
	y += 15
	setText(pdf, colorText)
	pdf.SetFont(fontName, "B", 26)
	centered(pdf, y, d.FullName)

	// Ism ostidagi ingichka chiziq (qo'lda to'ldirilgan blank taassuroti).
	// Siljish 26pt matn balandligidan katta bo'lishi shart, aks holda chiziq
	// ismning ustidan o'tib, uni chizib tashlagandek ko'rinadi.
	y += 12
	setDraw(pdf, colorMuted)
	pdf.SetLineWidth(0.2)
	pdf.Line(pageW/2-70, y, pageW/2+70, y)

	y += 8
	setText(pdf, colorMuted)
	pdf.SetFont(fontName, "", 11)
	centered(pdf, y, textFor)

	y += 11
	setText(pdf, colorPrimary)
	pdf.SetFont(fontName, "B", 17)
	centered(pdf, y, d.Title)

	// O'lchov (masalan "100 000 qadam") — bo'lsa aksent rangda.
	if d.ValueLabel != "" {
		y += 10
		setText(pdf, colorAccent)
		pdf.SetFont(fontName, "B", 13)
		centered(pdf, y, d.ValueLabel)
	}

	// Izoh: uzun bo'lsa bir necha qatorga bo'linadi.
	if sub := firstNonEmpty(d.Note, d.Description); sub != "" {
		y += 9
		setText(pdf, colorMuted)
		pdf.SetFont(fontName, "", 10)
		y = multiline(pdf, y, sub, pageW-2*(margin+30), 5)
	}

	return y
}

// drawFooter — sana, raqam va imzo joyi (sahifa pastiga qat'iy joylashadi).
func drawFooter(pdf *fpdf.Fpdf, d Data, _ float64) {
	y := pageH - margin - 30

	setDraw(pdf, colorMuted)
	pdf.SetLineWidth(0.2)

	// Chap: sana. O'ng: imzo chizig'i.
	leftX, rightX := margin+22.0, pageW-margin-22.0
	lineW := 55.0

	// Muhr — imzo chizig'ining chap uchiga qisman tushadi (rasmiy
	// hujjatlarda muhr odatda imzoga tegib turadi). Markazi chiziqdan
	// PASTDA: tepada sertifikat tavsifi bo'lishi mumkin.
	drawStamp(pdf, d.Signing, rightX-lineW+2, y+7)

	// Imzo skani chiziq USTIGA tushadi (chiziq — imzo uchun joy).
	// O'ngga surilgan: markazlashtirilgan tavsif matni sahifaning shu
	// balandligida chap tomonda tugaydi, ustma-ust tushmasin.
	if len(d.Signing.SignaturePNG) > 0 {
		pdf.ImageOptions(signatureName,
			rightX-signatureW, y-signatureH-0.5,
			signatureW, signatureH,
			false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}

	setDraw(pdf, colorMuted)
	pdf.SetLineWidth(0.2)
	pdf.Line(leftX, y, leftX+lineW, y)
	pdf.Line(rightX-lineW, y, rightX, y)

	setText(pdf, colorText)
	pdf.SetFont(fontName, "", 10)
	pdf.SetXY(leftX, y+1)
	pdf.CellFormat(lineW, 6, d.AwardedAt.Format("02.01.2006"), "", 0, "C", false, 0, "")
	// O'ng chiziq ostida imzolovchining ismi (bo'lsa).
	pdf.SetXY(rightX-lineW, y+1)
	pdf.CellFormat(lineW, 6, d.Signing.SignerName, "", 0, "C", false, 0, "")

	setText(pdf, colorMuted)
	pdf.SetFont(fontName, "", 8)
	pdf.SetXY(leftX, y+7)
	pdf.CellFormat(lineW, 5, textDateLabel, "", 0, "C", false, 0, "")
	pdf.SetXY(rightX-lineW, y+7)
	// Lavozim ko'rsatilmagan bo'lsa umumiy "Institut rahbariyati".
	pdf.CellFormat(lineW, 5, firstNonEmpty(d.Signing.SignerTitle, textSignature), "", 0, "C", false, 0, "")

	// Sertifikat raqami — eng pastda, markazda. Y ichki ramkadan yuqorida
	// tugashi kerak (ramka pastki chizig'i: pageH-margin-4), aks holda yozuv
	// ramkaga tegib ketadi.
	if d.Number != "" {
		pdf.SetFont(fontName, "", 8)
		centered(pdf, pageH-margin-14, fmt.Sprintf("%s № %s", textNumberLabel, d.Number))
	}
}

// drawStamp — muhr: haqiqiy skan yoki NAMUNA belgisi.
//
// (cx, cy) — muhr MARKAZI.
func drawStamp(pdf *fpdf.Fpdf, s Signing, cx, cy float64) {
	if len(s.StampPNG) > 0 {
		pdf.ImageOptions(stampName, cx-stampR, cy-stampR, 2*stampR, 2*stampR,
			false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		return
	}
	if !s.Sample {
		return
	}

	// NAMUNA muhri — ATAYLAB rasmiy muhrga o'xshamaydi.
	//
	// Institutning haqiqiy muhriga taqlid qilingan tasvir soxta rasmiy
	// hujjat yasashga yaraydi, shuning uchun bu yerda oddiy aylana va ochiq
	// "NAMUNA" yozuvi ishlatilgan: PDF ni ko'rgan har kim buni haqiqiy emas
	// deb darrov tushunadi. Haqiqiy muhr CERT_STAMP_PATH orqali qo'yiladi.
	setDraw(pdf, colorMuted)
	pdf.SetLineWidth(0.8)
	pdf.Circle(cx, cy, stampR, "D")
	pdf.SetLineWidth(0.3)
	pdf.Circle(cx, cy, stampR-2.5, "D")

	setText(pdf, colorMuted)
	pdf.SetFont(fontName, "B", 9)
	pdf.SetXY(cx-stampR, cy-4.5)
	pdf.CellFormat(2*stampR, 5, textSampleStamp, "", 0, "C", false, 0, "")

	pdf.SetFont(fontName, "", 4)
	pdf.SetXY(cx-stampR, cy+0.5)
	pdf.CellFormat(2*stampR, 3.5, textSampleStampNote, "", 0, "C", false, 0, "")
}

// ─────────────────────────── Yordamchilar ───────────────────────────

// centered — matnni sahifa bo'ylab markazlaydi (Y — matn tepasi).
func centered(pdf *fpdf.Fpdf, y float64, text string) {
	pdf.SetXY(margin, y)
	pdf.CellFormat(pageW-2*margin, 8, text, "", 0, "C", false, 0, "")
}

// multiline — uzun matnni belgilangan kenglikda qatorlarga bo'lib chizadi.
// Qaytaradi: oxirgi qatordan keyingi Y.
func multiline(pdf *fpdf.Fpdf, y float64, text string, width, lineH float64) float64 {
	for _, line := range pdf.SplitLines([]byte(text), width) {
		centered(pdf, y, string(line))
		y += lineH
	}
	return y
}

func setText(pdf *fpdf.Fpdf, c rgb) { pdf.SetTextColor(c.R, c.G, c.B) }
func setDraw(pdf *fpdf.Fpdf, c rgb) { pdf.SetDrawColor(c.R, c.G, c.B) }

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
