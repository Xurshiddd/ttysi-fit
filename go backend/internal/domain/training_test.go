package domain

import "testing"

// TestValidTrainingLevel — daraja enum: SQL filtriga tushadi, qat'iy bo'lsin.
func TestValidTrainingLevel(t *testing.T) {
	for _, l := range []string{TrainingBeginner, TrainingIntermediate, TrainingAdvanced} {
		if !ValidTrainingLevel(l) {
			t.Errorf("haqiqiy daraja rad etildi: %q", l)
		}
	}
	for _, l := range []string{"", "expert", "BEGINNER", "beginner' OR '1'='1"} {
		if ValidTrainingLevel(l) {
			t.Errorf("noto'g'ri daraja qabul qilindi: %q", l)
		}
	}
}

func TestValidTrainingStatus(t *testing.T) {
	if !ValidTrainingStatus("draft") || !ValidTrainingStatus("published") {
		t.Error("haqiqiy holat rad etildi")
	}
	for _, s := range []string{"", "active", "published; DROP TABLE trainings"} {
		if ValidTrainingStatus(s) {
			t.Errorf("noto'g'ri holat qabul qilindi: %q", s)
		}
	}
}

// TestTrainingLevelsAreDistinct — darajalar takrorlanmasin va bo'sh bo'lmasin.
func TestTrainingLevelsAreDistinct(t *testing.T) {
	seen := map[string]bool{}
	for _, l := range []string{TrainingBeginner, TrainingIntermediate, TrainingAdvanced} {
		if l == "" {
			t.Error("bo'sh daraja konstantasi")
		}
		if seen[l] {
			t.Errorf("takrorlangan daraja: %q", l)
		}
		seen[l] = true
	}
}
