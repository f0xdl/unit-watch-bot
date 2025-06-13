package tgservice

import "sync"

type State string

const (
	StateIdle       State = "idle"
	StateWaitingUID       = "waiting_uid"
)

type FSM struct {
	mu     sync.RWMutex
	states map[int64]State
}

func NewFSM() *FSM {
	return &FSM{states: make(map[int64]State)}
}

func (f *FSM) Get(chatID int64) State {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if s, ok := f.states[chatID]; ok {
		return s
	}
	return StateIdle
}

func (f *FSM) Set(chatID int64, state State) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.states[chatID] = state
}
