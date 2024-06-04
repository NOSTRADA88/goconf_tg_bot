// Package fsm provides a finite state machine for managing user states.
package fsm

import (
	"context"
	"errors"
	"github.com/NOSTRADA88/telegram-bot-go/internal/repository/redis"
)

// StateController interface defines the methods for getting and setting user states.
type StateController interface {
	// GetState retrieves the current state of a user.
	// It takes a context and a key representing the user ID.
	// It returns the current state as a string and an error if any occurred.
	GetState(ctx context.Context, key int64) (string, error)

	// SetState sets the state of a user.
	// It takes a context, a key representing the user ID, and the state to be set.
	// It returns an error if any occurred.
	SetState(ctx context.Context, key int64, state string) error
}

// FSM struct is a finite state machine that uses a Redis cache client for state management.
type FSM struct {
	// rdb is a Redis cache client used for state management.
	rdb redis.CacheClient
}

// New function creates a new FSM with a given Redis cache client.
// It takes a Redis cache client and returns a pointer to a new FSM.
func New(rdb redis.CacheClient) *FSM {
	return &FSM{rdb: rdb}
}

// GetState method retrieves the current state of a user.
// It takes a context and a key representing the user ID.
// It returns the current state as a string and an error if any occurred.
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

// SetState method sets the state of a user.
// It takes a context, a key representing the user ID, and the state to be set.
// It returns an error if any occurred.
func (fsm *FSM) SetState(ctx context.Context, key int64, state string) error {
	return fsm.rdb.Set(ctx, key, state, 0)
}
