package domain

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestMakeExcerpt — excerpt bo'sh bo'lsa body'dan yasaladi.
func TestMakeExcerpt(t *testing.T) {
	tests := []struct {
		name string
		body string
		max  int
		want string
	}{
		{"qisqa matn o'zgarmaydi", "Salom dunyo", 100, "Salom dunyo"},
		{"bo'sh matn", "", 100, ""},
		{"faqat bo'shliq", "   \n\t ", 100, ""},
		{
			"ortiqcha bo'shliqlar bitta bo'ladi",
			"Universitet   krossi\n\n  boshlandi",
			100,
			"Universitet krossi boshlandi",
		},
		{
			"uzun matn so'z chegarasida kesiladi",
			"Universitet krossi ertaga stadionda boshlanadi",
			20,
			"Universitet krossi…",
		},
		{"aniq chegarada kesilmaydi", "abcde", 5, "abcde"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeExcerpt(tt.body, tt.max); got != tt.want {
				t.Errorf("MakeExcerpt(%q, %d) = %q, kutilgan %q", tt.body, tt.max, got, tt.want)
			}
		})
	}
}

// TestMakeExcerpt_Unicode — kirill va o'zbek apostrofli harflar buzilmasligi kerak.
//
// Bayt bo'yicha kesish ko'p baytli harfni o'rtasidan bo'lib, buzuq belgi
// (U+FFFD) hosil qilardi. Rune bo'yicha kesish shuni oldini oladi.
func TestMakeExcerpt_Unicode(t *testing.T) {
	body := "Универсиада бошланди ва барча талабалар иштирок этмоқда бу ерда"
	got := MakeExcerpt(body, 20)

	if !utf8.ValidString(got) {
		t.Fatalf("natija buzuq UTF-8: %q", got)
	}
	if strings.ContainsRune(got, '�') {
		t.Errorf("buzuq belgi (U+FFFD) topildi: %q", got)
	}
	// Kesilgan bo'lishi kerak (asl matn uzunroq).
	if !strings.HasSuffix(got, "…") {
		t.Errorf("uzun matn kesilmadi: %q", got)
	}
	// Rune bo'yicha: natija maxRunes+1 (…) dan oshmasin.
	if n := utf8.RuneCountInString(got); n > 21 {
		t.Errorf("juda uzun: %d rune, %q", n, got)
	}
}

// TestMakeExcerpt_UzbekApostrophe — o'zbek matnidagi ' va ' belgilari.
func TestMakeExcerpt_UzbekApostrophe(t *testing.T) {
	body := "Bugun universitetda o‘quvchilar uchun katta sport musobaqasi bo‘lib o‘tdi"
	got := MakeExcerpt(body, 30)

	if !utf8.ValidString(got) {
		t.Fatalf("natija buzuq UTF-8: %q", got)
	}
	if strings.ContainsRune(got, '�') {
		t.Errorf("buzuq belgi topildi: %q", got)
	}
}

func TestValidNewsStatus(t *testing.T) {
	if !ValidNewsStatus("draft") || !ValidNewsStatus("published") {
		t.Error("haqiqiy holat rad etildi")
	}
	for _, s := range []string{"", "active", "PUBLISHED", "published'; DROP TABLE news;--"} {
		if ValidNewsStatus(s) {
			t.Errorf("noto'g'ri holat qabul qilindi: %q", s)
		}
	}
}
