package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	ZapLogger struct {
		Logger *zap.Logger
		level  zap.AtomicLevel
	}

	Config struct {
		Level string
	}
)

func New(cfg Config) (*ZapLogger, error) {

	var level zapcore.Level
	level.Set(cfg.Level)
	atomic := zap.NewAtomicLevelAt(level)
	settings := defaultSettings(atomic)

	l, err := settings.config.Build(settings.opts...)
	if err != nil {
		return nil, err
	}

	return &ZapLogger{
		Logger: l,
		level:  atomic,
	}, nil
}

func (z *ZapLogger) SetLevel(cfg Config) {
	var level zapcore.Level
	level.Set(cfg.Level)
	z.level.SetLevel(level)
}

func (z *ZapLogger) Start(ctx context.Context) error {
	return nil
}

func (z *ZapLogger) Stop(ctx context.Context) error {
	z.Logger.Sync()
	return nil
}
