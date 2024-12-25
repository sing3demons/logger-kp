package logger

import (
	"context"
	"log"
	"os"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const key = "logger"

func NewLogger(options ...zap.Option) *zap.Logger {
	zap.NewProduction()
	if configLog.AppLog.LogFile {
		if err := ensureLogDirExists(configLog.AppLog.Name); err != nil {
			log.Fatal(err)
		}
		return newLogFile(configLog.AppLog.Name)
	}

	encCfg := zapcore.EncoderConfig{
		MessageKey:   "msg",
		TimeKey:      "time",
		LevelKey:     "level",
		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	if configLog.AppLog.LogLevel == 0 {
		configLog.AppLog.LogLevel = zapcore.DebugLevel
	}

	// File encoder using console format
	consoleEncoder := zapcore.NewConsoleEncoder(encCfg)

	// Create a zapcore core
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), configLog.AppLog.LogLevel),
	)

	// Create logger
	log := zap.New(core, options...)
	return log

}

func NewLog(c context.Context) *zap.Logger {
	switch logger := c.Value(key).(type) {
	case *zap.Logger:
		return logger
	default:
		return zap.NewNop()
	}
}

func InitSession(c context.Context, logger *zap.Logger) *zap.Logger {
	// get session from context
	session := c.Value(xSession)
	if session == nil {
		uuidV7, err := uuid.NewV7()
		if err != nil {
			uuidV7 = uuid.New()
		}
		session = uuidV7.String()
		// set session to context
		c = context.WithValue(c, xSession, session)
	}

	// set session to logger
	l := logger.With(zap.String("session", c.Value(xSession).(string)))
	// set logger to context
	c = context.WithValue(c, key, l)
	return l
}
