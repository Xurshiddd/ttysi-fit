package handler

import "testing"

// TestAbsoluteMediaURL — DB dagi nisbiy yo'lga asos qo'shiladi, absolyut URL
// va bo'sh qiymat tegilmaydi. Bu chegara muhim: HEMIS avatarlari absolyut
// saqlanadi, o'zimiznikilar nisbiy — ikkalasi bitta ustunda yashaydi.
func TestAbsoluteMediaURL(t *testing.T) {
	const base = "http://localhost:8090"

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"nisbiy yo'lga asos qo'shiladi", "/static/avatars/1.jpg", "http://localhost:8090/static/avatars/1.jpg"},
		{"bo'sh qiymat o'zgarmaydi", "", ""},
		{"absolyut http tegilmaydi", "http://other.host/a.jpg", "http://other.host/a.jpg"},
		{"absolyut https (HEMIS) tegilmaydi", "https://hemis.ttyesi.uz/static/pi/5/d/x.jpg", "https://hemis.ttyesi.uz/static/pi/5/d/x.jpg"},
		{"slashsiz nisbiy yo'l absolyut deb qaraladi", "static/avatars/1.jpg", "static/avatars/1.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := absoluteMediaURL(base, tt.in); got != tt.want {
				t.Errorf("absoluteMediaURL(%q, %q) = %q, kutilgan %q", base, tt.in, got, tt.want)
			}
		})
	}
}

// TestAbsoluteMediaURL_BoshAsos — asos sozlanmagan bo'lsa nisbiy yo'l o'zgarishsiz
// qaytadi (mijoz uni o'z API asosiga nisbatan hal qiladi).
func TestAbsoluteMediaURL_BoshAsos(t *testing.T) {
	if got := absoluteMediaURL("", "/static/avatars/1.jpg"); got != "/static/avatars/1.jpg" {
		t.Errorf("bo'sh asosda yo'l o'zgardi: %q", got)
	}
}
