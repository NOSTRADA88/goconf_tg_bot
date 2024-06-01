package redis_test

import (
	"context"
	"github.com/NOSTRADA88/telegram-bot-go/internal/repository/redis"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"testing"
)

func NewMockedClient() (*redis.Client, redismock.ClientMock) {
	db, mock := redismock.NewClientMock()
	return &redis.Client{Rdb: db}, mock
}

func TestPingReturnsNoErrorWhenRedisIsAvailable(t *testing.T) {
	client, mock := NewMockedClient()
	mock.ExpectPing().SetVal("PONG")

	err := client.Ping(context.Background())
	assert.NoError(t, err)
}

func TestPingReturnsErrorWhenRedisIsUnavailable(t *testing.T) {
	client, mock := NewMockedClient()
	mock.ExpectPing().SetErr(redis.Nil)

	err := client.Ping(context.Background())
	assert.Error(t, err)
}

func TestGetReturnsValueWhenKeyExists(t *testing.T) {
	client, mock := NewMockedClient()
	mock.ExpectGet("1").SetVal("value")

	val, err := client.Get(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "value", val)
}

func TestGetReturnsErrorWhenKeyDoesNotExist(t *testing.T) {
	client, mock := NewMockedClient()
	mock.ExpectGet("1").SetErr(redis.Nil)

	val, err := client.Get(context.Background(), 1)
	assert.Error(t, err)
	assert.Empty(t, val)
}

func TestSetStoresValueWithNoError(t *testing.T) {
	client, mock := NewMockedClient()
	mock.ExpectSet("1", "value", 0).SetVal("OK")

	err := client.Set(context.Background(), 1, "value", 0)
	assert.NoError(t, err)
}

func TestSetReturnsErrorWhenCannotStoreValue(t *testing.T) {
	client, mock := NewMockedClient()
	mock.ExpectSet("1", "value", 0).SetErr(redis.Nil)

	err := client.Set(context.Background(), 1, "value", 0)
	assert.Error(t, err)
}
