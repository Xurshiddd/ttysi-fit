package domain

import "testing"

// TestValidateChallengeConfig — turga qarab config validatsiyasi (§16.2).
// Bu qoidalar chellenjning yakunlanishini belgilaydi: config noto'g'ri bo'lsa,
// maqsad hech qachon hisoblanmaydi va chellenj abadiy ochiq qoladi.
func TestValidateChallengeConfig(t *testing.T) {
	tests := []struct {
		name    string
		typ     ChallengeType
		config  string
		wantErr bool
	}{
		{"steps: to'g'ri", ChallengeSteps, `{"target_steps":10000}`, false},
		{"steps: maqsad yo'q", ChallengeSteps, `{}`, true},
		{"steps: nol maqsad", ChallengeSteps, `{"target_steps":0}`, true},
		{"steps: manfiy maqsad", ChallengeSteps, `{"target_steps":-5}`, true},
		{"steps: matn maqsad", ChallengeSteps, `{"target_steps":"10000"}`, true},

		// Eng muhim holat: xato yozilgan kalit jimgina saqlanmasligi kerak.
		{"steps: noma'lum kalit rad etiladi", ChallengeSteps, `{"target_step":10000}`, true},
		{"steps: ortiqcha kalit rad etiladi", ChallengeSteps, `{"target_steps":100,"hack":1}`, true},

		{"distance: to'g'ri", ChallengeDistance, `{"target_km":42.2}`, false},
		{"distance: juda kichik", ChallengeDistance, `{"target_km":0.05}`, true},
		{"active_min: to'g'ri", ChallengeActiveMin, `{"target_min":30}`, false},

		{"custom: bo'sh config o'tadi", ChallengeCustom, `{}`, false},
		{"custom: izoh bilan", ChallengeCustom, `{"note":"Bahorgi aksiya"}`, false},
		{"custom: izoh raqam bo'lsa rad", ChallengeCustom, `{"note":123}`, true},

		{"noma'lum tur", ChallengeType("yoq"), `{}`, true},
		{"buzuq JSON", ChallengeSteps, `{oops`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChallengeConfig(tt.typ, []byte(tt.config))
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateChallengeConfig(%q, %s) xato=%v, kutilgan xato=%v",
					tt.typ, tt.config, err, tt.wantErr)
			}
		})
	}
}

// TestChallengeTarget — config'dagi maqsad metrika birligiga o'giriladi.
// distance uchun km -> metr: `activities.distance_m` metrda saqlanadi, agar
// o'girish bo'lmasa 42 km maqsad 42 metrda "bajarilgan" bo'lib qolardi.
func TestChallengeTarget(t *testing.T) {
	tests := []struct {
		name   string
		typ    ChallengeType
		config string
		want   float64
	}{
		{"steps birliksiz", ChallengeSteps, `{"target_steps":10000}`, 10000},
		{"distance km -> metr", ChallengeDistance, `{"target_km":42.2}`, 42200},
		{"active_min birliksiz", ChallengeActiveMin, `{"target_min":30}`, 30},
		{"custom maqsadsiz", ChallengeCustom, `{"note":"x"}`, 0},
		{"maqsad yo'q", ChallengeSteps, `{}`, 0},
		{"buzuq JSON qulamaydi", ChallengeSteps, `{oops`, 0},
		{"noma'lum tur", ChallengeType("yoq"), `{"target_steps":1}`, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ChallengeTarget(tt.typ, []byte(tt.config)); got != tt.want {
				t.Errorf("ChallengeTarget(%q, %s) = %v, kutilgan %v", tt.typ, tt.config, got, tt.want)
			}
		})
	}
}

// TestChallengeTypeSpecs — registr admin formasi uchun to'liq ta'rif beradi.
func TestChallengeTypeSpecs(t *testing.T) {
	specs := ChallengeTypeSpecs()
	if len(specs) == 0 {
		t.Fatal("tur ro'yxati bo'sh")
	}
	for _, s := range specs {
		if s.Label == "" {
			t.Errorf("%q turining yorlig'i yo'q", s.Type)
		}
		if !ValidChallengeType(string(s.Type)) {
			t.Errorf("%q registrda yo'q", s.Type)
		}
		// Metrikali turda maqsad kaliti bo'lishi shart, aks holda progress
		// hisoblansa ham hech qachon yakunlanmaydi.
		if s.Metric != "" && s.TargetKey == "" {
			t.Errorf("%q: metrika bor, maqsad kaliti yo'q", s.Type)
		}
		// Maqsad kaliti registrdagi maydonlar ichida bo'lishi kerak.
		if s.TargetKey != "" {
			found := false
			for _, f := range s.Fields {
				if f.Key == s.TargetKey {
					found = true
				}
			}
			if !found {
				t.Errorf("%q: TargetKey %q maydonlar ichida yo'q", s.Type, s.TargetKey)
			}
		}
	}
}

// TestValidChallengeEnums — enum tekshiruvlari SQL'ga qiymat o'tkazishdan oldin
// ishlaydi, shuning uchun ular qat'iy bo'lishi kerak.
func TestValidChallengeEnums(t *testing.T) {
	if !ValidChallengeStatus("active") || !ValidChallengeStatus("draft") || !ValidChallengeStatus("finished") {
		t.Error("haqiqiy holat rad etildi")
	}
	if ValidChallengeStatus("") || ValidChallengeStatus("DROP TABLE") {
		t.Error("noto'g'ri holat qabul qilindi")
	}
	if !ValidChallengeScope("university") || ValidChallengeScope("world") {
		t.Error("qamrov tekshiruvi noto'g'ri")
	}
	if !ValidChallengeType("steps") || ValidChallengeType("steps; DROP") {
		t.Error("tur tekshiruvi noto'g'ri")
	}
}
