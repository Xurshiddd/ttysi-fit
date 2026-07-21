// Package report — hisobotlarni faylga (CSV) yozish yordamchilari.
package report

import (
	"encoding/csv"
	"io"
	"strings"
)

// utf8BOM — Excel uchun zarur bayt ketma-ketligi (U+FEFF).
//
// BOM'siz Excel CSV'ni tizim kod sahifasida o'qiydi va o'zbek lotin/kirill
// harflari ("O'zbekiston", "Ҳисобот") krakozyabra bo'lib chiqadi. LibreOffice
// va Google Sheets BOM bilan ham to'g'ri ochadi.
//
// Baytlar sifatida yozilgan: manba faylning o'rtasidagi BOM belgisini Go
// kompilyatori qabul qilmaydi.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Writer — xavfsiz CSV yozuvchi.
type Writer struct {
	w *csv.Writer
}

// NewWriter — BOM yozib, CSV yozuvchini tayyorlaydi.
//
// Ajratgich sifatida nuqtali vergul (;) ishlatiladi: o'zbek/rus lokalidagi
// Excel'da ro'yxat ajratgichi aynan shu, vergul bo'lsa butun qator bitta
// katakka tushib qoladi.
func NewWriter(w io.Writer) *Writer {
	_, _ = w.Write(utf8BOM)
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	return &Writer{w: cw}
}

// Write — bitta qator yozadi. Har bir katak formula injection'dan tozalanadi.
func (x *Writer) Write(fields ...string) error {
	safe := make([]string, len(fields))
	for i, f := range fields {
		safe[i] = Sanitize(f)
	}
	return x.w.Write(safe)
}

// Flush — buferni chiqaradi va yig'ilgan xatoni qaytaradi.
func (x *Writer) Flush() error {
	x.w.Flush()
	return x.w.Error()
}

// Sanitize — CSV formula injection (CSV injection) himoyasi.
//
// Excel/LibreOffice `=`, `+`, `-`, `@` bilan boshlangan katakni FORMULA deb
// hisoblaydi. Ismi "=cmd|'/c calc'!A1" bo'lgan foydalanuvchi yaratilsa,
// hisobotni ochgan admin kompyuterida buyruq bajarilishi mumkin edi
// (CLAUDE.md §17.3 — injection). Bunday katak oldiga apostrof qo'yiladi:
// Excel uni oddiy matn sifatida ko'rsatadi.
//
// Tab va CR ham tekshiriladi: ular ham ba'zi versiyalarda formula
// boshlanishini yashirishga ishlatiladi.
func Sanitize(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + clean(s)
	}
	return clean(s)
}

// clean — qator ichidagi boshqaruv belgilarini olib tashlaydi (CSV
// tuzilishini buzmasin). Yangi qator probelga almashtiriladi.
func clean(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r':
			return ' '
		}
		if r < 0x20 && r != '\t' {
			return -1
		}
		return r
	}, s)
}
