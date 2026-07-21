package validation

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// e164Re — E.164 telefon formati: + va 2..15 raqam (birinchisi 0 emas).
var e164Re = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// RegisterJSONTagNames — gin validatorini sozlaydi: maydon xatolarida
// Go struct nomi o'rniga `json` tag nomi ishlatilsin (masalan "full_name").
// Shu orqali i18n maydon yorliqlari to'g'ri moslanadi.
func RegisterJSONTagNames() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// RegisterCustomRules — loyihaga xos validatsiya qoidalari.
func RegisterCustomRules() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	// e164_opt — bo'sh satr YOKI E.164 telefon.
	//
	// Nega kerak: pointer maydonda `omitempty` faqat nil ni o'tkazadi. Ya'ni
	// {"phone":""} (raqamni tozalash niyati) bo'sh satr sifatida `e164` ga
	// tushib rad etilardi va foydalanuvchi raqamini o'chira olmasdi.
	// Bu qoida "tegilmasin" (nil), "tozalansin" ("") va "yangi raqam"ni ajratadi.
	_ = v.RegisterValidation("e164_opt", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		return s == "" || e164Re.MatchString(s)
	})
}
