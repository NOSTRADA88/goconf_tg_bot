package fsm

import (
	"context"
	"errors"
	"github.com/NOSTRADA88/telegram-bot-go/internal/repository/redis"
)

type StateController interface {
	GetState(ctx context.Context, key int64) (string, error)
	SetState(ctx context.Context, key int64, state string) error
}

// FSM struct implements the simplest finite state machine. Used redis to save data from reloads
type FSM struct {
	rdb redis.CacheClient
}

// New function creates new FSM pointer
func New(rdb redis.CacheClient) *FSM {
	return &FSM{rdb: rdb}
}

// GetState method gets current user state. State depends on pressed buttons in telegram
func (fsm *FSM) GetState(ctx context.Context, key int64) (string, error) {
	state, err := fsm.rdb.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return state, nil
		}
		return "", err
	}
	return state, nil
}

// SetState method sets user state. State depends on pressed buttons in telegram
func (fsm *FSM) SetState(ctx context.Context, key int64, state string) error {
	return fsm.rdb.Set(ctx, key, state, 0)
}
