package certificate

import (
	"bytes"
	"image/png"
	"strings"
	"testing"
	"time"

	"golang.org/x/image/font/sfnt"
)

// TestRender — sertifikat chizila oladimi va natija haqiqiy PDF mi.
func TestRender(t *testing.T) {
	out, err := Render(Data{
		FullName:   "To'rayev Zafarjon O'ktamovich",
		Title:      "Birinchi ming qadam",
		ValueLabel: "1 000 qadam",
		Number:     "463DDE23",
		AwardedAt:  time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Errorf("natija PDF emas, boshi: %q", out[:min(8, len(out))])
	}
	if len(out) < 2000 {
		t.Errorf("PDF juda kichik (%d bayt) — shrift yoki kontent tushib qolgan", len(out))
	}
}

// TestRenderDefaults — bo'sh maydonlar standart qiymatga tushadi va
// generatsiya qulamaydi. Sertifikat ixtiyoriy maydonlarsiz ham chizilishi
// kerak: qo'lda berilgan yutuqda ValueLabel bo'lmaydi.
func TestRenderDefaults(t *testing.T) {
	out, err := Render(Data{FullName: "Aziz", Title: "Ishtirok uchun"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Error("minimal ma'lumot bilan PDF chizilmadi")
	}
}

// TestRenderUzunMatn — uzun izoh sahifadan chiqib ketmasligi kerak.
// SetAutoPageBreak(false) tufayli u ikkinchi sahifa ochmaydi; bu yerda
// generatsiya qulamasligi tekshiriladi.
func TestRenderUzunMatn(t *testing.T) {
	long := strings.Repeat("Bahorgi universitet krossida faol ishtirok etgani uchun. ", 12)
	out, err := Render(Data{
		FullName:    "Test Foydalanuvchi",
		Title:       "Universitet krossi g'olibi",
		Description: long,
		Note:        long,
	})
	if err != nil {
		t.Fatalf("Render (uzun matn): %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Error("uzun matnda PDF chizilmadi")
	}
}

// TestShriftOzbekHarflariniQoplaydi — REGRESSIYA QO'RIQCHISI.
//
// Go'ning o'z shrifti (gofont) o'zbek kirillchasidagi Қ, Ғ, Ҳ va lotin ʻ
// harflarini qoplamaydi — ular PDF'da BO'SH JOY bo'lib chiqadi va buni faqat
// tayyor faylni ochib ko'rgandagina sezish mumkin. Shrift almashtirilsa, bu
// test darrov ogohlantiradi.
func TestShriftOzbekHarflariniQoplaydi(t *testing.T) {
	// HEMIS ma'lumotlarida uchraydigan harflar: o'zbek lotin + o'zbek kirill.
	required := []rune{
		'ʻ', '‘', '’', // lotin: oʻ, gʻ
		'Ў', 'ў', 'Қ', 'қ', 'Ғ', 'ғ', 'Ҳ', 'ҳ', // o'zbek kirill
		'Й', 'й', 'Ц', 'щ', 'э', 'ю', 'я', // umumiy kirill
		'№', '–', // hujjatda uchraydigan belgilar
	}

	for name, data := range map[string][]byte{
		"regular": fontRegular,
		"bold":    fontBold,
	} {
		f, err := sfnt.Parse(data)
		if err != nil {
			t.Fatalf("%s shriftni o'qib bo'lmadi: %v", name, err)
		}
		var b sfnt.Buffer
		for _, r := range required {
			idx, err := f.GlyphIndex(&b, r)
			if err != nil {
				t.Errorf("%s: %q (U+%04X) qidirishda xato: %v", name, r, r, err)
				continue
			}
			if idx == 0 {
				t.Errorf("%s shriftda %q (U+%04X) YO'Q — sertifikatda bo'sh joy chiqadi",
					name, r, r)
			}
		}
	}
}

// TestLogoAktivi — emblema PNG binarga to'g'ri singdirilganmi.
//
// NEGA: `go:embed` fayl yo'qolsa kompilyatsiyada xato beradi, LEKIN fayl
// buzilgan yoki noto'g'ri formatda bo'lsa xato faqat PDF chizilganda chiqadi.
// Alfa kanali alohida tekshiriladi: u bo'lsa fpdf shaffof piksellarni QORA
// qilib chizadi va emblema atrofida kulrang to'rtburchak paydo bo'ladi
// (assets/README.md ga qarang).
func TestLogoAktivi(t *testing.T) {
	if len(logoPNG) == 0 {
		t.Fatal("emblema PNG bo'sh")
	}
	if !bytes.HasPrefix(logoPNG, []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatal("emblema fayli PNG emas")
	}

	cfg, err := png.DecodeConfig(bytes.NewReader(logoPNG))
	if err != nil {
		t.Fatalf("emblemani o'qib bo'lmadi: %v", err)
	}
	if cfg.Width < 300 || cfg.Height < 300 {
		t.Errorf("emblema juda kichik (%dx%d) — chop etishda xira chiqadi",
			cfg.Width, cfg.Height)
	}

	// Alfa kanalini IHDR dagi rang turidan aniqlaymiz.
	//
	// png.DecodeConfig BUNGA YARAMAYDI: u alfasiz truecolor PNG uchun ham
	// color.RGBAModel qaytaradi, ya'ni model bo'yicha alfa borligini ajratib
	// bo'lmaydi.
	//
	// PNG tuzilishi: imzo(8) + uzunlik(4) + "IHDR"(4) + kenglik(4) +
	// balandlik(4) + bit chuqurligi(1) + RANG TURI(1) -> 25-bayt.
	// Rang turlari: 0=kulrang, 2=truecolor, 3=palitra, 4=kulrang+alfa,
	// 6=truecolor+alfa.
	const colorTypeOffset = 25
	if len(logoPNG) <= colorTypeOffset {
		t.Fatal("PNG juda qisqa — IHDR o'qib bo'lmadi")
	}
	if ct := logoPNG[colorTypeOffset]; ct == 4 || ct == 6 {
		t.Errorf("emblemada alfa kanali bor (rang turi %d) — fpdf shaffof "+
			"joylarni qora qilib chizadi va emblema atrofida kulrang "+
			"to'rtburchak paydo bo'ladi. assets/README.md: alpha=False bilan "+
			"qayta yarating", ct)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
