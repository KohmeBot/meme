package meme

import (
	"sync"
)

type Tasks struct {
	rw    sync.RWMutex
	tasks map[int64]bool
}

func NewTasks() *Tasks {
	return &Tasks{tasks: make(map[int64]bool)}
}

func (t *Tasks) AddTask(uid int64) bool {
	t.rw.Lock()
	defer t.rw.Unlock()
	b, ok := t.tasks[uid]
	if !ok || !b {
		t.tasks[uid] = true
	}
	return !b
}

func (t *Tasks) Done(uid int64) {
	t.rw.Lock()
	defer t.rw.Unlock()
	t.tasks[uid] = false
}
