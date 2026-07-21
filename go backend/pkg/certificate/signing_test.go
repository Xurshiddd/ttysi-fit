package certificate

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writePNG — sinov uchun PNG yasaydi. withAlpha=true bo'lsa alfa kanali bilan.
func writePNG(t *testing.T, name string, withAlpha bool) string {
	t.Helper()

	var img image.Image
	if withAlpha {
		m := image.NewRGBA(image.Rect(0, 0, 8, 8))
		m.Set(0, 0, color.RGBA{255, 0, 0, 128})
		img = m
	} else {
		m := image.NewNRGBA(image.Rect(0, 0, 8, 8))
		for x := 0; x < 8; x++ {
			for y := 0; y < 8; y++ {
				m.Set(x, y, color.RGBA{10, 20, 30, 255})
			}
		}
		// NRGBA ham alfa bilan yoziladi — alfasiz uchun paletта/gray kerak.
		img = &opaque{m}
	}

	path := filepath.Join(t.TempDir(), name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("fayl yaratilmadi: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("PNG yozilmadi: %v", err)
	}
	return path
}

// opaque — image.Image ni alfasiz (RGB) qilib ko'rsatadi, shunda png.Encode
// rang turi 2 (truecolor) bilan yozadi.
type opaque struct{ src image.Image }

func (o *opaque) ColorModel() color.Model { return color.RGBAModel }
func (o *opaque) Bounds() image.Rectangle { return o.src.Bounds() }
func (o *opaque) At(x, y int) color.Color {
	r, g, b, _ := o.src.At(x, y).RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}
}
func (o *opaque) Opaque() bool { return true }

// Alfa kanali bor PNG rad etilishi kerak.
//
// NEGA MUHIM: fpdf shaffof piksellarni QORA qilib chizadi — muhr atrofida
// qora to'rtburchak paydo bo'ladi. Bu xato faqat tayyor PDF ochilganda
// ko'rinardi, ya'ni allaqachon foydalanuvchiga ketgan bo'lardi.
func TestLoadSigning_RejectsAlphaPNG(t *testing.T) {
	path := writePNG(t, "muhr.png", true)

	_, err := LoadSigning(path, "", "", "", false)
	if err == nil {
		t.Fatal("alfa kanali bor PNG qabul qilindi")
	}
	if !strings.Contains(err.Error(), "alfa") {
		t.Errorf("xato sababi tushunarsiz: %v", err)
	}
}

func TestLoadSigning_AcceptsOpaquePNG(t *testing.T) {
	path := writePNG(t, "muhr.png", false)

	s, err := LoadSigning(path, "", "Rektor A.", "Rektor", false)
	if err != nil {
		t.Fatalf("yaroqli PNG rad etildi: %v", err)
	}
	if len(s.StampPNG) == 0 {
		t.Error("muhr yuklanmadi")
	}
	if s.SignerName != "Rektor A." || s.SignerTitle != "Rektor" {
		t.Error("imzolovchi ma'lumoti saqlanmadi")
	}
}

// PNG bo'lmagan fayl rad etilsin (kimdir JPG qo'yib qo'ysa).
func TestLoadSigning_RejectsNonPNG(t *testing.T) {
	path := filepath.Join(t.TempDir(), "muhr.png")
	if err := os.WriteFile(path, []byte("bu PNG emas"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadSigning(path, "", "", "", false); err == nil {
		t.Fatal("PNG bo'lmagan fayl qabul qilindi")
	}
}

// Yo'q fayl — aniq xato bilan (jimgina muhrsiz qolmasin).
func TestLoadSigning_MissingFileFails(t *testing.T) {
	if _, err := LoadSigning("/yoq/muhr.png", "", "", "", false); err == nil {
		t.Fatal("mavjud bo'lmagan fayl qabul qilindi")
	}
}

// Yo'l ko'rsatilmasa — xato emas: muhr hali skanerlanmagan bo'lsa ham
// sertifikat berilaverishi kerak.
func TestLoadSigning_EmptyPathsAreOK(t *testing.T) {
	s, err := LoadSigning("", "", "", "", false)
	if err != nil {
		t.Fatalf("bo'sh yo'llar xato berdi: %v", err)
	}
	if s.HasStamp() {
		t.Error("muhrsiz sozlamada HasStamp true qaytdi")
	}
}

// Haqiqiy muhr berilsa NAMUNA o'chirilishi kerak — ikkalasi birga chizilmasin.
func TestLoadSigning_RealStampDisablesSample(t *testing.T) {
	path := writePNG(t, "muhr.png", false)

	s, err := LoadSigning(path, "", "", "", true)
	if err != nil {
		t.Fatalf("LoadSigning: %v", err)
	}
	if s.Sample {
		t.Error("haqiqiy muhr bor, lekin NAMUNA ham yoqilgan qoldi")
	}
}

// NAMUNA muhri bilan PDF chizilsin va "NAMUNA" yozuvi PDF ichida bo'lsin.
//
// NEGA: bu belgi hujjat haqiqiy emasligini bildiradi. Agar u jimgina
// tushib qolsa, sinov sertifikati rasmiy hujjatdek ko'rinib qolardi.
func TestRender_SampleStampIsVisible(t *testing.T) {
	pdf, err := Render(Data{
		FullName:  "Toshmatov Alisher",
		Title:     "Birinchi ming qadam",
		Number:    "ABCD1234",
		AwardedAt: time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC),
		Signing:   Signing{Sample: true},
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(pdf) == 0 {
		t.Fatal("PDF bo'sh")
	}
	// fpdf UTF-8 matnni siqilgan oqimda saqlaydi, shuning uchun baytdan
	// izlash ishonchsiz. O'rniga: NAMUNA'siz va NAMUNA bilan chizilgan
	// PDF hajmi FARQ qilishi kerak (qo'shimcha aylana va matn).
	plain, err := Render(Data{
		FullName:  "Toshmatov Alisher",
		Title:     "Birinchi ming qadam",
		Number:    "ABCD1234",
		AwardedAt: time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Render (namunasiz): %v", err)
	}
	if len(pdf) <= len(plain) {
		t.Errorf("NAMUNA muhri chizilmagan ko'rinadi (%d <= %d bayt)", len(pdf), len(plain))
	}
}

// Imzolovchi ismi/lavozimi berilsa PDF o'zgarishi kerak.
func TestRender_SignerAppears(t *testing.T) {
	base := Data{FullName: "Test", Title: "Yutuq", AwardedAt: time.Now()}

	withSigner := base
	withSigner.Signing = Signing{SignerName: "A. A. Aliyev", SignerTitle: "Rektor"}

	a, err := Render(base)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Render(withSigner)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(a, b) {
		t.Error("imzolovchi ma'lumoti PDF ga tushmadi")
	}
}

// Muhrsiz ham sertifikat chizilishi kerak (hozirgi xatti-harakat buzilmasin).
func TestRender_WorksWithoutSigning(t *testing.T) {
	pdf, err := Render(Data{FullName: "Test", Title: "Yutuq", AwardedAt: time.Now()})
	if err != nil {
		t.Fatalf("muhrsiz Render: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF")) {
		t.Error("natija PDF emas")
	}
}
