package domain

import (
	"testing"
	"time"
)

// TestCompetitionRegOpen — ro'yxatdan o'tish shartlari.
// Bu qoida ikki joyda ishlatiladi (yozishdan oldin tekshirish va UI tugmasi),
// shuning uchun uni bir joyda va aniq sinash muhim.
func TestCompetitionRegOpen(t *testing.T) {
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)
	max5 := 5
	unlimited := 0

	tests := []struct {
		name         string
		status       string
		regEndsAt    *time.Time
		maxPart      *int
		participants int
		want         bool
	}{
		{"ochiq: muddatsiz, cheklovsiz", CompStatusRegistration, nil, nil, 0, true},
		{"ochiq: muddat kelajakda", CompStatusRegistration, &future, nil, 0, true},
		{"ochiq: joy bor", CompStatusRegistration, nil, &max5, 4, true},

		{"yopiq: draft", CompStatusDraft, nil, nil, 0, false},
		{"yopiq: ongoing", CompStatusOngoing, nil, nil, 0, false},
		{"yopiq: finished", CompStatusFinished, nil, nil, 0, false},
		{"yopiq: muddat o'tgan", CompStatusRegistration, &past, nil, 0, false},
		{"yopiq: joy to'lgan", CompStatusRegistration, nil, &max5, 5, false},
		{"yopiq: joydan oshgan", CompStatusRegistration, nil, &max5, 6, false},

		// 0 — cheklovsiz degani, "joy yo'q" emas.
		{"ochiq: max=0 cheklovsiz", CompStatusRegistration, nil, &unlimited, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Competition{
				Status:          tt.status,
				RegEndsAt:       tt.regEndsAt,
				MaxParticipants: tt.maxPart,
			}
			if got := c.RegOpen(tt.participants, now); got != tt.want {
				t.Errorf("RegOpen() = %v, kutilgan %v", got, tt.want)
			}
		})
	}
}

// TestValidateCompetitionConfig — turga qarab config validatsiyasi (§16.3).
// Chellenj bilan bir xil `validateFields` ni ishlatadi, shuning uchun bu yerda
// asosan MUSOBAQAGA XOS turlar va maydonlar tekshiriladi.
func TestValidateCompetitionConfig(t *testing.T) {
	tests := []struct {
		name    string
		typ     CompetitionType
		config  string
		wantErr bool
	}{
		{"individual: to'g'ri", CompetitionIndividual, `{"sport":"yugurish"}`, false},
		{"individual: sport yo'q", CompetitionIndividual, `{}`, true},
		{"individual: sport raqam", CompetitionIndividual, `{"sport":5}`, true},

		{"team: to'g'ri", CompetitionTeam, `{"sport":"futbol","team_size":5}`, false},
		{"team: jamoa a'zosi yo'q", CompetitionTeam, `{"sport":"futbol"}`, true},
		{"team: bitta kishilik jamoa rad etiladi", CompetitionTeam, `{"sport":"futbol","team_size":1}`, true},

		{"faculty_vs: to'g'ri metrika", CompetitionFacultyVs, `{"sport":"yurish","metric":"steps"}`, false},
		{"faculty_vs: noma'lum metrika", CompetitionFacultyVs, `{"sport":"yurish","metric":"hack"}`, true},
		{"faculty_vs: metrika yo'q", CompetitionFacultyVs, `{"sport":"yurish"}`, true},

		{"custom: bo'sh config", CompetitionCustom, `{}`, false},

		// Registrda yo'q kalit jimgina saqlanib qolmasligi kerak.
		{"noma'lum maydon rad etiladi", CompetitionIndividual, `{"sport":"a","hack":1}`, true},
		{"noma'lum tur", CompetitionType("yoq"), `{}`, true},
		{"buzuq JSON", CompetitionIndividual, `{oops`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCompetitionConfig(tt.typ, []byte(tt.config))
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCompetitionConfig(%q, %s) xato=%v, kutilgan xato=%v",
					tt.typ, tt.config, err, tt.wantErr)
			}
		})
	}
}

// TestValidCompetitionEnums — enum qiymatlari SQL'ga o'tishdan oldin tekshiriladi.
func TestValidCompetitionEnums(t *testing.T) {
	for _, s := range []string{CompStatusDraft, CompStatusRegistration, CompStatusOngoing, CompStatusFinished} {
		if !ValidCompetitionStatus(s) {
			t.Errorf("haqiqiy holat rad etildi: %q", s)
		}
	}
	for _, s := range []string{"", "active", "DROP TABLE competitions"} {
		if ValidCompetitionStatus(s) {
			t.Errorf("noto'g'ri holat qabul qilindi: %q", s)
		}
	}
	if !ValidCompetitionType("individual") || ValidCompetitionType("individual'; --") {
		t.Error("tur tekshiruvi noto'g'ri")
	}
}

// TestCompetitionTypeSpecs — registr admin formasi uchun to'liq ta'rif beradi.
func TestCompetitionTypeSpecs(t *testing.T) {
	specs := CompetitionTypeSpecs()
	if len(specs) == 0 {
		t.Fatal("tur ro'yxati bo'sh")
	}
	for _, s := range specs {
		if s.Label == "" {
			t.Errorf("%q turining yorlig'i yo'q", s.Type)
		}
		if !ValidCompetitionType(string(s.Type)) {
			t.Errorf("%q registrda yo'q", s.Type)
		}
		// Select maydonda variantlar bo'lishi shart, aks holda admin hech
		// qanday qiymat tanlay olmaydi va forma boshi berk ko'cha bo'ladi.
		for _, f := range s.Fields {
			if f.Type == FieldSelect && len(f.Options) == 0 {
				t.Errorf("%q.%s: select maydonda variant yo'q", s.Type, f.Key)
			}
		}
	}
}
