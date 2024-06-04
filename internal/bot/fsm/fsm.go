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
	// It takes a key representing the user ID.
	// It returns the current state as a string and an error if any occurred.
	GetState(key int64) (string, error)

	// SetState sets the state of a user.
	// It takes a key representing the user ID, and the state to be set.
	// It returns an error if any occurred.
	SetState(key int64, state string) error
}

// FSM struct is a finite state machine that uses a Redis cache client for state management.
// rdb is a Redis cache client used for state management.
// ctx is the context in which the FSM operates.
type FSM struct {
	rdb redis.CacheClient
	ctx context.Context
}

// New function creates a new FSM with a given Redis cache client.
// It takes a Redis cache client and a context and returns a pointer to a new FSM.
func New(rdb redis.CacheClient, ctx context.Context) *FSM {
	return &FSM{rdb: rdb, ctx: ctx}
}

// GetState method retrieves the current state of a user.
// It takes a key representing the user ID.
// It returns the current state as a string and an error if any occurred.
func (fsm *FSM) GetState(key int64) (string, error) {
	state, err := fsm.rdb.Get(fsm.ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return state, nil
		}
		return "", err
	}
	return state, nil
}

// SetState method sets the state of a user.
// It takes a key representing the user ID, and the state to be set.
// It returns an error if any occurred.
func (fsm *FSM) SetState(key int64, state string) error {
	return fsm.rdb.Set(fsm.ctx, key, state, 0)
}
