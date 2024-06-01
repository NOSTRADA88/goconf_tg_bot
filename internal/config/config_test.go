package config_test

import (
	"github.com/NOSTRADA88/telegram-bot-go/internal/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDefault(t *testing.T) {
	cfg, err := config.New()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 27017, cfg.Database.Port)

	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)

}
