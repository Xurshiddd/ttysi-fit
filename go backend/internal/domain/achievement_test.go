package domain

import "testing"

// TestValidateAchievementCriteria — mezon turga qarab tekshiriladi (§16.2).
// Bu qoidalar yutuqning berilishini belgilaydi: mezon noto'g'ri bo'lsa,
// maqsad 0 bo'lib qoladi va avtomatik yutuq hech qachon berilmaydi.
func TestValidateAchievementCriteria(t *testing.T) {
	tests := []struct {
		name     string
		typ      AchievementType
		criteria string
		wantErr  bool
	}{
		{"steps_total: to'g'ri", AchStepsTotal, `{"threshold":100000}`, false},
		{"steps_total: mezon yo'q", AchStepsTotal, `{}`, true},
		{"steps_total: nol", AchStepsTotal, `{"threshold":0}`, true},
		{"steps_total: manfiy", AchStepsTotal, `{"threshold":-5}`, true},
		{"steps_total: matn", AchStepsTotal, `{"threshold":"100000"}`, true},

		// Eng muhim holat: xato yozilgan kalit jimgina saqlanmasligi kerak,
		// aks holda yutuq hech qachon berilmaydi va sababi ko'rinmaydi.
		{"steps_total: noma'lum kalit rad etiladi", AchStepsTotal, `{"treshold":100000}`, true},
		{"steps_total: ortiqcha kalit rad etiladi", AchStepsTotal, `{"threshold":10,"hack":1}`, true},

		{"distance_total: to'g'ri", AchDistanceTotal, `{"threshold":42.2}`, false},
		{"distance_total: juda kichik", AchDistanceTotal, `{"threshold":0.05}`, true},
		{"active_days: to'g'ri", AchActiveDays, `{"threshold":30}`, false},
		{"challenge_count: to'g'ri", AchChallengeCount, `{"threshold":5}`, false},

		{"manual: bo'sh mezon o'tadi", AchManual, `{}`, false},
		{"manual: izoh bilan", AchManual, `{"note":"Universitet krossi g'olibi"}`, false},
		{"manual: izoh raqam bo'lsa rad", AchManual, `{"note":123}`, true},

		{"noma'lum tur", AchievementType("yoq"), `{}`, true},
		{"buzuq JSON", AchStepsTotal, `{oops`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAchievementCriteria(tt.typ, []byte(tt.criteria))
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAchievementCriteria(%q, %s) xato=%v, kutilgan xato=%v",
					tt.typ, tt.criteria, err, tt.wantErr)
			}
		})
	}
}

// TestAchievementThreshold — mezon metrika birligiga o'giriladi.
// distance uchun km -> metr: `activities.distance_m` metrda saqlanadi, agar
// o'girish bo'lmasa 42 km mezon 42 metrda "bajarilgan" bo'lib qolardi.
func TestAchievementThreshold(t *testing.T) {
	tests := []struct {
		name     string
		typ      AchievementType
		criteria string
		want     float64
	}{
		{"steps_total", AchStepsTotal, `{"threshold":100000}`, 100000},
		{"distance_total km -> metr", AchDistanceTotal, `{"threshold":42}`, 42000},
		{"active_days", AchActiveDays, `{"threshold":30}`, 30},
		{"challenge_count", AchChallengeCount, `{"threshold":5}`, 5},

		// Maqsadsiz va buzuq holatlar 0 qaytaradi — chaqiruvchi 0 ni
		// "avtomatik berilmaydi" deb tushunadi.
		{"manual maqsadsiz", AchManual, `{"note":"x"}`, 0},
		{"bo'sh mezon", AchStepsTotal, `{}`, 0},
		{"buzuq JSON", AchStepsTotal, `{oops`, 0},
		{"noma'lum tur", AchievementType("yoq"), `{"threshold":10}`, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AchievementThreshold(tt.typ, []byte(tt.criteria))
			if got != tt.want {
				t.Errorf("AchievementThreshold(%q, %s) = %v, kutilgan %v",
					tt.typ, tt.criteria, got, tt.want)
			}
		})
	}
}

// TestAchievementProgressPct — foiz 0..100 oralig'idan chiqmasligi kerak:
// mezondan oshib ketgan progress mobil progress-barni buzib yubormasin.
func TestAchievementProgressPct(t *testing.T) {
	tests := []struct {
		name     string
		progress float64
		target   float64
		want     float64
	}{
		{"yarmi", 50, 100, 50},
		{"to'liq", 100, 100, 100},
		{"oshib ketgan 100 da qoladi", 250, 100, 100},
		{"maqsadsiz (manual) 0", 50, 0, 0},
		{"manfiy maqsad 0", 50, -10, 0},
		{"manfiy progress 0", -5, 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AchievementProgressPct(tt.progress, tt.target)
			if got != tt.want {
				t.Errorf("AchievementProgressPct(%v, %v) = %v, kutilgan %v",
					tt.progress, tt.target, got, tt.want)
			}
		})
	}
}

// TestAchievementValueLabel — sertifikatga chiqadigan o'lchov matni.
// Masofa metrda saqlanadi: km ga o'girilmasa sertifikatda "42200 km" chiqardi.
func TestAchievementValueLabel(t *testing.T) {
	tests := []struct {
		name  string
		typ   AchievementType
		value float64
		want  string
	}{
		{"qadam guruhlanadi", AchStepsTotal, 4238, "4 238 qadam"},
		{"katta qadam", AchStepsTotal, 1000000, "1 000 000 qadam"},
		{"kichik qadam guruhlanmaydi", AchStepsTotal, 999, "999 qadam"},
		{"masofa metr -> km", AchDistanceTotal, 42200, "42.2 km"},
		{"butun km da .0 tushadi", AchDistanceTotal, 5000, "5 km"},
		{"faol kunlar", AchActiveDays, 30, "30 kun"},
		{"chellenjlar", AchChallengeCount, 5, "5 ta chellenj"},

		// Qo'lda beriladigan yutuqda o'lchov yo'q — sertifikatda bu qator
		// umuman chizilmaydi.
		{"manual bo'sh", AchManual, 1, ""},
		{"nol qiymat bo'sh", AchStepsTotal, 0, ""},
		{"manfiy qiymat bo'sh", AchStepsTotal, -5, ""},
		{"noma'lum tur bo'sh", AchievementType("yoq"), 100, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AchievementValueLabel(tt.typ, tt.value)
			if got != tt.want {
				t.Errorf("AchievementValueLabel(%q, %v) = %q, kutilgan %q",
					tt.typ, tt.value, got, tt.want)
			}
		})
	}
}

// TestAchievementAwardMode — berish usuli TURDAN kelib chiqadi.
//
// Bu xavfsizlik qoidasi (§17.3 №26/27 ruhida): agar admin avtomatik yutuqni
// "manual" qilib belgilay olsa, mezonni butunlay chetlab o'tib, xohlagan
// odamga yutuq bera olardi va reyting adolati buzilardi.
func TestAchievementAwardMode(t *testing.T) {
	auto := []AchievementType{AchStepsTotal, AchDistanceTotal, AchActiveDays, AchChallengeCount}
	for _, typ := range auto {
		spec, ok := AchievementSpec(typ)
		if !ok {
			t.Fatalf("%q registrda yo'q", typ)
		}
		if spec.AwardMode != AwardModeAuto {
			t.Errorf("%q award_mode = %q, kutilgan %q", typ, spec.AwardMode, AwardModeAuto)
		}
		if spec.Source == "" {
			t.Errorf("%q avtomatik, lekin Source bo'sh — progress hisoblanmaydi", typ)
		}
	}

	spec, ok := AchievementSpec(AchManual)
	if !ok {
		t.Fatal("manual tur registrda yo'q")
	}
	if spec.AwardMode != AwardModeManual {
		t.Errorf("manual award_mode = %q, kutilgan %q", spec.AwardMode, AwardModeManual)
	}
	if spec.Source != "" {
		t.Errorf("manual tur Source ga ega (%q) — u avtomatik baholanmasligi kerak", spec.Source)
	}
}

// TestAchievementTypeSpecsBarqaror — registr va ro'yxat mos bo'lishi kerak.
// Yangi tur qo'shilib, tartib ro'yxatiga kiritilmasa, u admin panel formasida
// ko'rinmay qolardi va sababi hech qayerda bilinmasdi.
func TestAchievementTypeSpecsBarqaror(t *testing.T) {
	specs := AchievementTypeSpecs()
	if len(specs) != len(achievementTypes) {
		t.Fatalf("AchievementTypeSpecs() %d ta qaytardi, registrda %d ta tur bor "+
			"— yangi tur tartib ro'yxatiga qo'shilmagan",
			len(specs), len(achievementTypes))
	}
	for _, s := range specs {
		if s.Type == "" {
			t.Error("spec turi bo'sh — registr kaliti bilan mos emas")
		}
		if !ValidAchievementType(string(s.Type)) {
			t.Errorf("%q ValidAchievementType dan o'tmadi", s.Type)
		}
	}
}

// TestAchievementMetricColumn — SQL'ga tushadigan ustun nomi faqat registrdan
// keladi (§3.2). Noma'lum tur uchun bo'sh qaytishi shart: aks holda repository
// bo'sh ustun nomi bilan so'rov yasab qo'yardi.
func TestAchievementMetricColumn(t *testing.T) {
	if got := AchievementMetricColumn(AchStepsTotal); got != "steps" {
		t.Errorf("steps_total ustuni = %q, kutilgan %q", got, "steps")
	}
	if got := AchievementMetricColumn(AchDistanceTotal); got != "distance_m" {
		t.Errorf("distance_total ustuni = %q, kutilgan %q", got, "distance_m")
	}
	for _, typ := range []AchievementType{AchActiveDays, AchChallengeCount, AchManual, "yoq"} {
		if got := AchievementMetricColumn(typ); got != "" {
			t.Errorf("%q uchun ustun %q qaytdi, bo'sh bo'lishi kerak", typ, got)
		}
	}
}
