package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logDir — kunlik log fayllari saqlanadigan papka.
const logDir = "logs"

// New — log darajasi va muhitga qarab zap logger yaratadi.
// Konsolga (inson o'qiy oladigan) va kunlik JSON faylga (logs/<sana>.log) yozadi.
func New(level, env string) (*zap.Logger, error) {
	lvl := zapcore.InfoLevel
	if l, err := zapcore.ParseLevel(level); err == nil {
		lvl = l
	}

	// Konsol core
	var consoleEnc zapcore.Encoder
	if env == "production" {
		consoleEnc = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	} else {
		encCfg := zap.NewDevelopmentEncoderConfig()
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		consoleEnc = zapcore.NewConsoleEncoder(encCfg)
	}
	consoleCore := zapcore.NewCore(consoleEnc, zapcore.Lock(os.Stdout), lvl)

	cores := []zapcore.Core{consoleCore}

	// Kunlik fayl core (xato bo'lsa ham konsol ishlashda davom etadi)
	if fileCore, err := dailyFileCore(lvl); err == nil {
		cores = append(cores, fileCore)
	}

	return zap.New(zapcore.NewTee(cores...), zap.AddCaller()), nil
}

// dailyFileCore — logs/<YYYY-MM-DD>.log fayliga yozuvchi JSON core.
func dailyFileCore(lvl zapcore.Level) (zapcore.Core, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}
	name := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	enc := zapcore.NewJSONEncoder(encCfg)
	return zapcore.NewCore(enc, zapcore.AddSync(f), lvl), nil
}
