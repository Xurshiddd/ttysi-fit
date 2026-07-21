package certificate

import (
	"bytes"
	"fmt"
	"os"
)

// Signing — sertifikatning imzo bloki: muhr, imzo skani va imzolovchi.
//
// Server ishga tushganda BIR MARTA yuklanadi (LoadSigning) va har chizishda
// qayta ishlatiladi: har so'rovda diskdan o'qish ham sekin, ham keraksiz.
type Signing struct {
	// StampPNG — muhr tasviri. Bo'sh bo'lsa muhr chizilmaydi.
	StampPNG []byte
	// SignaturePNG — imzo tasviri. Bo'sh bo'lsa faqat chiziq qoladi.
	SignaturePNG []byte

	// SignerName — imzolovchining F.I.O. si (masalan "A. A. Aliyev").
	SignerName string
	// SignerTitle — lavozimi (masalan "Rektor"). Bo'sh bo'lsa textSignature.
	SignerTitle string

	// Sample — NAMUNA muhrini chizish.
	//
	// Bu HAQIQIY muhr emas: ochiq "NAMUNA" yozuvi bilan vektor shaklida
	// chiziladi, shuning uchun rasmiy hujjat sifatida ishlatib bo'lmaydi.
	// Faqat ishlab chiqish/sinov uchun — config uni production'da yoqishga
	// yo'l qo'ymaydi (CLAUDE.md §17: yumshatish faqat APP_ENV=local da).
	Sample bool
}

// HasStamp — muhr (haqiqiy yoki namuna) chiziladimi.
func (s Signing) HasStamp() bool { return len(s.StampPNG) > 0 || s.Sample }

// LoadSigning — config'da ko'rsatilgan fayllarni o'qib tekshiradi.
//
// Yo'l bo'sh bo'lsa o'sha element shunchaki chizilmaydi (xato emas):
// muhr hali skanerlanmagan bo'lsa ham sertifikat berilaverishi kerak.
func LoadSigning(stampPath, signaturePath, signerName, signerTitle string, sample bool) (Signing, error) {
	s := Signing{
		SignerName:  signerName,
		SignerTitle: signerTitle,
		Sample:      sample,
	}

	var err error
	if stampPath != "" {
		if s.StampPNG, err = loadPNG(stampPath, "muhr"); err != nil {
			return Signing{}, err
		}
		// Haqiqiy muhr bor — namuna kerak emas.
		s.Sample = false
	}
	if signaturePath != "" {
		if s.SignaturePNG, err = loadPNG(signaturePath, "imzo"); err != nil {
			return Signing{}, err
		}
	}
	return s, nil
}

// loadPNG — PNG faylni o'qib, fpdf uchun yaroqliligini tekshiradi.
//
// Tekshiruv ATAYLAB qattiq va startupda bajariladi: alfa kanali bor PNG
// jimgina noto'g'ri chiziladi (shaffof piksellar QORA bo'ladi — muhr
// atrofida qora to'rtburchak paydo bo'ladi). Bu xato faqat tayyor PDF
// ochilganda ko'rinardi, ya'ni allaqachon foydalanuvchiga ketgan bo'lardi.
// Shuning uchun server umuman ishga tushmagani yaxshi.
func loadPNG(path, what string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("certificate: %s faylini o'qib bo'lmadi (%s): %w", what, path, err)
	}
	if !bytes.HasPrefix(b, []byte("\x89PNG\r\n\x1a\n")) {
		return nil, fmt.Errorf("certificate: %s fayli PNG emas: %s", what, path)
	}

	// Rang turi IHDR ning 25-baytida: 4 = kulrang+alfa, 6 = truecolor+alfa.
	// (png.DecodeConfig bunga yaramaydi — alfasiz truecolor uchun ham
	// RGBAModel qaytaradi; batafsil izoh certificate_test.go da.)
	const colorTypeOffset = 25
	if len(b) <= colorTypeOffset {
		return nil, fmt.Errorf("certificate: %s fayli buzuq (IHDR yo'q): %s", what, path)
	}
	if ct := b[colorTypeOffset]; ct == 4 || ct == 6 {
		return nil, fmt.Errorf(
			"certificate: %s faylida alfa kanali bor (rang turi %d): %s — fpdf "+
				"shaffof piksellarni QORA qilib chizadi. Faylni oq fonga "+
				"yassilang (ImageMagick: convert in.png -background white "+
				"-alpha remove -alpha off out.png)", what, ct, path)
	}
	return b, nil
}
