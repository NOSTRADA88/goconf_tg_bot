package fsm_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NOSTRADA88/telegram-bot-go/internal/bot/fsm"
	"github.com/NOSTRADA88/telegram-bot-go/internal/repository/redis"
)

type MockCacheClient struct {
	state string
	err   error
}

func (m *MockCacheClient) Get(ctx context.Context, key int64) (string, error) {
	return m.state, m.err
}

func (m *MockCacheClient) Set(ctx context.Context, key int64, value interface{}, duration time.Duration) error {
	return m.err
}

func TestGetStateReturnsStateWhenNoError(t *testing.T) {
	mockCacheClient := &MockCacheClient{state: "state", err: nil}

	f := fsm.New(mockCacheClient, context.Background())
	state, err := f.GetState(1)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if state != "state" {
		t.Errorf("Expected state 'state', got '%s'", state)
	}
}

func TestGetStateReturnsEmptyStringWhenRedisNil(t *testing.T) {
	mockCacheClient := &MockCacheClient{state: "", err: redis.Nil}

	f := fsm.New(mockCacheClient, context.Background())
	state, err := f.GetState(1)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if state != "" {
		t.Errorf("Expected state '', got '%s'", state)
	}
}

func TestGetStateReturnsErrorWhenRedisError(t *testing.T) {
	mockCacheClient := &MockCacheClient{state: "", err: errors.New("redis error")}

	f := fsm.New(mockCacheClient, context.Background())
	_, err := f.GetState(1)

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestSetStateReturnsNoErrorWhenSuccessful(t *testing.T) {
	mockCacheClient := &MockCacheClient{err: nil}

	f := fsm.New(mockCacheClient, context.Background())
	err := f.SetState(1, "state")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestSetStateReturnsErrorWhenUnsuccessful(t *testing.T) {
	mockCacheClient := &MockCacheClient{err: errors.New("redis error")}

	f := fsm.New(mockCacheClient, context.Background())
	err := f.SetState(1, "state")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
