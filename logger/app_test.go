package logger

import (
	"context"
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewLogger(t *testing.T) {
	// Test case 1: Log to file
	configLog = LogConfig{
		AppLog: AppLog{
			Name:    "test_app_test_log",
			LogFile: true,
		},
	}

	logger := NewLogger()
	if logger == nil {
		t.Fatal("expected logger to be created, but got nil")
	}

	// Check if log directory was created
	if _, err := os.Stat(configLog.AppLog.Name); os.IsNotExist(err) {
		t.Errorf("Expected log directory %s to be created, but it does not exist", configLog.AppLog.Name)
	}

	// Clean up
	os.RemoveAll(configLog.AppLog.Name)

	// Test case 2: Log to console
	configLog = LogConfig{
		AppLog: AppLog{
			LogFile:    false,
			LogConsole: true,
			LogLevel:   zapcore.DebugLevel,
		},
	}

	logger = NewLogger()
	if logger == nil {
		t.Fatal("Expected logger to be created, but got nil")
	}

	// Test case 3: Log with custom options
	configLog = LogConfig{
		AppLog: AppLog{
			LogFile:    false,
			LogConsole: true,
			LogLevel:   zapcore.DebugLevel,
		},
	}

	logger = NewLogger(zap.AddCaller())
	if logger == nil {
		t.Fatal("Expected logger to be created, but got nil")
	}
}
func TestNewLog(t *testing.T) {
	// Test case 1: Logger exists in context
	logger := zap.NewExample()
	ctx := context.WithValue(context.Background(), key, logger)
	retrievedLogger := NewLog(ctx)
	if retrievedLogger == nil {
		t.Fatal("Expected logger to be retrieved from context, but got nil")
	}

	// Test case 2: Logger does not exist in context
	ctx = context.Background()
	retrievedLogger = NewLog(ctx)
	if retrievedLogger == nil {
		t.Fatal("Expected a no-op logger to be returned, but got nil")
	}
}

func TestInitSession(t *testing.T) {
	// Test case 1: Session does not exist in context
	logger := zap.NewExample()
	ctx := context.Background()
	ctx, newLogger := InitSession(ctx, logger)
	if newLogger == nil {
		t.Fatal("Expected logger to be initialized with session, but got nil")
	}

	session := ctx.Value(xSession)
	if session == nil {
		t.Fatal("Expected session to be set in context, but got nil")
	}

	// Test case 2: Session already exists in context
	existingSession := "existing-session-id"
	ctx = context.WithValue(context.Background(), xSession, existingSession)
	ctx, newLogger = InitSession(ctx, logger)
	if newLogger == nil {
		t.Fatal("Expected logger to be initialized with existing session, but got nil")
	}

	retrievedSession := ctx.Value(xSession)
	if retrievedSession != existingSession {
		t.Fatalf("Expected session to be %s, but got %s", existingSession, retrievedSession)
	}
}
