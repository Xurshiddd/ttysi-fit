package i18n

import (
	"sort"
	"strconv"
	"strings"
)

// Locale — qo'llab-quvvatlanadigan til.
type Locale string

const (
	UZ      Locale = "uz"
	RU      Locale = "ru"
	EN      Locale = "en"
	Default Locale = UZ
)

var supported = map[Locale]struct{}{
	UZ: {}, RU: {}, EN: {},
}

// IsSupported — til qo'llab-quvvatlanishini tekshiradi.
func IsSupported(l Locale) bool {
	_, ok := supported[l]
	return ok
}

// Parse — bitta til tegini ("ru", "ru-RU", "RU") Locale ga aylantiradi.
func Parse(s string) (Locale, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if i := strings.IndexByte(s, '-'); i > 0 {
		s = s[:i]
	}
	l := Locale(s)
	if IsSupported(l) {
		return l, true
	}
	return Default, false
}

// ParseAcceptLanguage — "ru-RU,ru;q=0.9,en;q=0.8" headerdan eng mos tilni tanlaydi.
func ParseAcceptLanguage(header string) Locale {
	if header == "" {
		return Default
	}

	type entry struct {
		loc Locale
		q   float64
	}
	var entries []entry

	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		tag := part
		q := 1.0
		if i := strings.Index(part, ";q="); i >= 0 {
			tag = strings.TrimSpace(part[:i])
			if v, err := strconv.ParseFloat(part[i+3:], 64); err == nil {
				q = v
			}
		}
		if loc, ok := Parse(tag); ok {
			entries = append(entries, entry{loc, q})
		}
	}

	if len(entries) == 0 {
		return Default
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].q > entries[j].q
	})
	return entries[0].loc
}
