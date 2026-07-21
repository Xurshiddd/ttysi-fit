package validation

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// TestE164Opt — "tegilmasin" (nil) / "tozalansin" ("") / "yangi raqam" holatlari.
// Bo'sh satr o'tishi SHART: aks holda foydalanuvchi telefon raqamini o'chira olmaydi.
func TestE164Opt(t *testing.T) {
	RegisterCustomRules()

	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		t.Fatal("gin validator engine olinmadi")
	}

	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{"to'g'ri O'zbekiston raqami", "+998901234567", false},
		{"bo'sh satr — tozalash", "", false},
		{"+ siz rad etiladi", "901234567", true},
		{"harf bilan rad etiladi", "+99890abc4567", true},
		{"noldan boshlansa rad etiladi", "+0998901234567", true},
		{"juda uzun rad etiladi", "+9989012345678901234", true},
		{"faqat + rad etiladi", "+", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Var(tt.phone, "omitempty,e164_opt")
			if (err != nil) != tt.wantErr {
				t.Errorf("e164_opt(%q): xato=%v, kutilgan xato=%v", tt.phone, err, tt.wantErr)
			}
		})
	}
}
