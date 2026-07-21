package config

import (
	"strings"
	"testing"
	"time"
)

// APP_TIMEZONE yuklanishi va DSN ga tushishi kerak.
//
// NEGA: kunlik faollik chegarasi (qaysi qadam qaysi kunga tegishli) shu
// mintaqada hisoblanadi. UTC qolib ketsa O'zbekistonda (UTC+5) mahalliy
// 00:00–05:00 dagi qadamlar kechagi kunga yozilib, o'sha kunning yozuvini
// buzardi. DSN dagi TimeZone esa SQL tomondagi CURRENT_DATE ni (reyting
// davri, statistika) bir xil kunga bog'laydi.
func TestLoad_Timezone(t *testing.T) {
	t.Setenv("APP_TIMEZONE", "Asia/Tashkent")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("config yuklanmadi: %v", err)
	}
	if cfg.App.Timezone == nil {
		t.Fatal("App.Timezone nil qoldi")
	}
	if got := cfg.App.Timezone.String(); got != "Asia/Tashkent" {
		t.Errorf("mintaqa: kutilgan Asia/Tashkent, olingan %s", got)
	}
	if !strings.Contains(cfg.DB.DSN(), "TimeZone=Asia/Tashkent") {
		t.Errorf("DSN da TimeZone yo'q: %s", cfg.DB.DSN())
	}

	// Mintaqa haqiqatan UTC+5 ekanini tasdiqlaymiz — tzdata binarga
	// singdirilgani (import _ "time/tzdata") shu yerda tekshiriladi:
	// ansiz konteynerda LoadLocation jim UTC bermasin.
	_, offset := time.Date(2026, 7, 21, 12, 0, 0, 0, cfg.App.Timezone).Zone()
	if offset != 5*3600 {
		t.Errorf("UTC+5 kutilgandi, olingan %d soniya", offset)
	}
}

// Noto'g'ri mintaqa nomi — server ishga tushmasin (jim UTC ga qaytmasin).
func TestLoad_InvalidTimezoneFails(t *testing.T) {
	t.Setenv("APP_TIMEZONE", "Yoq/Mintaqa")

	if _, err := Load(""); err == nil {
		t.Fatal("noto'g'ri APP_TIMEZONE da xato kutilgandi")
	}
}

// "NAMUNA" muhri production'da taqiqlangan bo'lishi kerak.
//
// NEGA: bu sinov belgisi. Production'ga o'tib ketsa institut o'z
// talabalariga "haqiqiy muhr emas" yozuvi bosilgan sertifikat berardi.
// Muhrsiz chiqqani shundan yaxshi.
func TestValidate_SampleStampForbiddenInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("CERT_SAMPLE_STAMP", "true")
	// Production validate'idan o'tish uchun qolgan shartlar.
	t.Setenv("JWT_ACCESS_SECRET", strings.Repeat("a", 40))
	t.Setenv("JWT_REFRESH_SECRET", strings.Repeat("b", 40))
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("RATE_LIMIT_ENABLED", "true")

	_, err := Load("")
	if err == nil {
		t.Fatal("production da CERT_SAMPLE_STAMP=true qabul qilindi")
	}
	if !strings.Contains(err.Error(), "CERT_SAMPLE_STAMP") {
		t.Errorf("xato sababi noaniq: %v", err)
	}
}

// Local'da esa default yoqilgan bo'lsin — ishlab chiqishda qulaylik.
func TestCert_SampleStampDefaultsOnInLocal(t *testing.T) {
	t.Setenv("APP_ENV", "local")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("config yuklanmadi: %v", err)
	}
	if !cfg.Cert.SampleStamp {
		t.Error("local'da NAMUNA muhri default o'chiq qolibdi")
	}
}

// Default — .env da ko'rsatilmasa ham Toshkent (loyiha O'zbekistonda).
func TestLoad_TimezoneDefault(t *testing.T) {
	t.Setenv("APP_TIMEZONE", "")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("config yuklanmadi: %v", err)
	}
	if got := cfg.App.TimezoneName; got != "Asia/Tashkent" {
		t.Errorf("default: kutilgan Asia/Tashkent, olingan %s", got)
	}
}
