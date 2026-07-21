package domain

import "testing"

// TestValidCoinReason — `reason` DB'ga matn sifatida tushadi va tarix filtri
// orqali SQL'ga boradi, shuning uchun ro'yxat qat'iy bo'lishi kerak.
func TestValidCoinReason(t *testing.T) {
	valid := []string{
		CoinReasonChallengeReward,
		CoinReasonCompetitionReward,
		CoinReasonAdminGrant,
		CoinReasonAdminRevoke,
	}
	for _, r := range valid {
		if !ValidCoinReason(r) {
			t.Errorf("haqiqiy sabab rad etildi: %q", r)
		}
	}

	invalid := []string{"", "hack", "admin_grant; DROP TABLE fit_coins", "ADMIN_GRANT"}
	for _, r := range invalid {
		if ValidCoinReason(r) {
			t.Errorf("noto'g'ri sabab qabul qilindi: %q", r)
		}
	}
}

// TestCoinReasonsAreDistinct — konstantalar bir-biriga to'qnashmasligi kerak:
// aks holda idempotentlik indeksi (user_id, reason, ref_id) noto'g'ri ishlaydi.
func TestCoinReasonsAreDistinct(t *testing.T) {
	all := []string{
		CoinReasonChallengeReward,
		CoinReasonCompetitionReward,
		CoinReasonAdminGrant,
		CoinReasonAdminRevoke,
	}
	seen := map[string]bool{}
	for _, r := range all {
		if r == "" {
			t.Error("bo'sh sabab konstantasi")
		}
		if seen[r] {
			t.Errorf("takrorlangan sabab: %q", r)
		}
		seen[r] = true
	}
}
