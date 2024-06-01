package logger_test

import (
	"github.com/NOSTRADA88/telegram-bot-go/internal/logger"
	"testing"
)

func TestInfo(t *testing.T) {
	log := logger.New(logger.InfoLevel)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked")
		}
	}()

	log.Info("test message")
}

func TestWarn(t *testing.T) {
	log := logger.New(logger.WarnLevel)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked")
		}
	}()

	log.Warn("test message")
}

func TestError(t *testing.T) {
	log := logger.New(logger.ErrorLevel)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked")
		}
	}()

	log.Error("test message")
}

func TestDebug(t *testing.T) {
	log := logger.New(logger.DebugLevel)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked")
		}
	}()

	log.Debug("test message")
}

func TestInfoF(t *testing.T) {
	log := logger.New(logger.InfoLevel)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked")
		}
	}()

	log.InfoF("%s %s", "test", "message")
}

func TestWarnF(t *testing.T) {
	log := logger.New(logger.WarnLevel)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked")
		}
	}()

	log.WarnF("%s %s", "test", "message")
}

func TestErrorF(t *testing.T) {
	log := logger.New(logger.ErrorLevel)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked")
		}
	}()

	log.ErrorF("%s %s", "test", "message")
}

func TestDebugF(t *testing.T) {
	log := logger.New(logger.DebugLevel)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked")
		}
	}()

	log.DebugF("%s %s", "test", "message")
}
