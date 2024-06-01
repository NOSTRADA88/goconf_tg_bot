package fsm_test

import (
	"context"
	"errors"
	"github.com/NOSTRADA88/telegram-bot-go/internal/bot/fsm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockCacheClient struct {
	mock.Mock
}

func (m *MockCacheClient) Get(ctx context.Context, key int64) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheClient) Set(ctx context.Context, key int64, value interface{}, duration time.Duration) error {
	args := m.Called(ctx, key, value, duration)
	return args.Error(0)
}

func (m *MockCacheClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestGetStateReturnsCorrectState(t *testing.T) {
	mockCacheClient := new(MockCacheClient)
	mockCacheClient.On("Get", mock.Anything, int64(1)).Return("state", nil)

	f := fsm.New(mockCacheClient)
	state, err := f.GetState(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, "state", state)
}

func TestGetStateReturnsErrorWhenGetFails(t *testing.T) {
	mockCacheClient := new(MockCacheClient)
	mockCacheClient.On("Get", mock.Anything, int64(1)).Return("", errors.New("get failed"))

	f := fsm.New(mockCacheClient)
	_, err := f.GetState(context.Background(), 1)

	assert.Error(t, err)
}

func TestSetStateReturnsNoErrorWhenSetSucceeds(t *testing.T) {
	mockCacheClient := new(MockCacheClient)
	mockCacheClient.On("Set", mock.Anything, int64(1), "state", mock.AnythingOfType("time.Duration")).Return(nil)

	f := fsm.New(mockCacheClient)
	err := f.SetState(context.Background(), 1, "state")

	assert.NoError(t, err)
}

func TestSetStateReturnsErrorWhenSetFails(t *testing.T) {
	mockCacheClient := new(MockCacheClient)
	mockCacheClient.On("Set", mock.Anything, int64(1), "state", mock.AnythingOfType("time.Duration")).Return(errors.New("set failed"))

	f := fsm.New(mockCacheClient)
	err := f.SetState(context.Background(), 1, "state")

	assert.Error(t, err)
}
