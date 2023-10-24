package foundation

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// globalLogger is global zap logger.
	globalLogger Logging = NewQuietLogger("info")
)

type Logging interface {
	With(args ...interface{}) Logging
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
}

func Logger() Logging {
	return globalLogger
}

func NewQuietLogger(lvlStr string) *QuietLogger {
	// First, define our level-handling logic.
	globalLevel, err := zapcore.ParseLevel(lvlStr)
	if err != nil {
		log.Fatalf("failed to initialize global logger: %v", err)
	}

	// High-priority output should also go to standard error, and low-priority
	// output should also go to standard out.
	// It is useful for Kubernetes deployment.
	// Kubernetes interprets os.Stdout log items as INFO and os.Stderr log items
	// as ERROR by default.
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= globalLevel && lvl < zapcore.ErrorLevel
	})
	consoleInfos := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	// Configure console output.
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder
	consoleEncoder := zapcore.NewJSONEncoder(cfg)

	// Join the outputs, encoders, and level-handling functions into zapcore
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleInfos, lowPriority),
	)

	// From a zapcore.Core, it's easy to construct a Logger.
	zapLogger := zap.New(core)
	zap.RedirectStdLog(zapLogger)

	return &QuietLogger{
		internal: zapLogger.Sugar(),
	}
}

type QuietLogger struct {
	internal *zap.SugaredLogger
}

func (s *QuietLogger) With(args ...interface{}) Logging {
	return &QuietLogger{
		internal: s.internal.With(args...),
	}
}

func (s *QuietLogger) Debug(args ...interface{}) {
	s.internal.Debug(args...)
}

func (s *QuietLogger) Info(args ...interface{}) {
	s.internal.Info(args...)
}

func (s *QuietLogger) Warn(args ...interface{}) {
	s.internal.Warn(args...)
}

func (s *QuietLogger) Error(args ...interface{}) {
	s.internal.Error(args...)
}

func (s *QuietLogger) Fatal(args ...interface{}) {
	s.internal.Fatal(args...)
}

func (s *QuietLogger) Debugf(template string, args ...interface{}) {
	s.internal.Debugf(template, args...)
}

func (s *QuietLogger) Infof(template string, args ...interface{}) {
	s.internal.Infof(template, args...)
}

func (s *QuietLogger) Warnf(template string, args ...interface{}) {
	s.internal.Warnf(template, args...)
}

func (s *QuietLogger) Errorf(template string, args ...interface{}) {
	s.internal.Errorf(template, args...)
}

func (s *QuietLogger) Fatalf(template string, args ...interface{}) {
	s.internal.Fatalf(template, args...)
}
