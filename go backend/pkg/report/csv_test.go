package report

import (
	"bytes"
	"strings"
	"testing"
)

// CSV formula injection — eng muhim tekshiruv.
//
// NEGA: hisobotdagi ism/email foydalanuvchi kiritgan matn. Excel `=` bilan
// boshlangan katakni formula deb bajaradi, ya'ni hisobotni ochgan admin
// hujum qurboni bo'lardi (§17.3 — injection).
func TestSanitize_FormulaInjection(t *testing.T) {
	cases := []struct{ in, want string }{
		{`=cmd|'/c calc'!A1`, `'=cmd|'/c calc'!A1`},
		{`+1+1`, `'+1+1`},
		{`-2+3`, `'-2+3`},
		{`@SUM(A1:A9)`, `'@SUM(A1:A9)`},
		{"\t=1+1", "'\t=1+1"},
		{"\r=1+1", "' =1+1"}, // CR probelga aylanadi
		// Oddiy qiymatlar o'zgarmasligi kerak.
		{"Toshmatov Alisher", "Toshmatov Alisher"},
		{"student@ttyesi.uz", "student@ttyesi.uz"},
		{"Iqtisodiyot fakulteti", "Iqtisodiyot fakulteti"},
		{"", ""},
	}
	for _, c := range cases {
		if got := Sanitize(c.in); got != c.want {
			t.Errorf("Sanitize(%q) = %q, kutilgan %q", c.in, got, c.want)
		}
	}
}

// Yangi qator katak ichida CSV tuzilishini buzmasligi kerak.
func TestSanitize_StripsNewlines(t *testing.T) {
	got := Sanitize("Birinchi\nIkkinchi")
	if strings.ContainsAny(got, "\n\r") {
		t.Errorf("yangi qator qolib ketdi: %q", got)
	}
	if got != "Birinchi Ikkinchi" {
		t.Errorf("olingan %q", got)
	}
}

// BOM bo'lmasa Excel o'zbek harflarini buzadi.
func TestWriter_WritesBOM(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.Write("Ism", "Fakultet"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	out := buf.Bytes()
	if len(out) < 3 || out[0] != 0xEF || out[1] != 0xBB || out[2] != 0xBF {
		t.Error("UTF-8 BOM yozilmadi")
	}
}

// Ajratgich nuqtali vergul — o'zbek/rus lokalidagi Excel uchun.
func TestWriter_UsesSemicolon(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	_ = w.Write("a", "b")
	_ = w.Flush()

	if !strings.Contains(buf.String(), "a;b") {
		t.Errorf("nuqtali vergul ishlatilmadi: %q", buf.String())
	}
}

// Yozilgan qatorda ham sanitize ishlashi kerak (faqat Sanitize'da emas).
func TestWriter_SanitizesRows(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	_ = w.Write("=1+1", "normal")
	_ = w.Flush()

	if !strings.Contains(buf.String(), `'=1+1`) {
		t.Errorf("qator tozalanmadi: %q", buf.String())
	}
}
