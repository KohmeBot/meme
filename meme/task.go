package meme

import (
	"sync"
	"time"
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

type TaskDuration struct {
	dur time.Duration
	rw  sync.RWMutex
	tMp map[int64]time.Time
}

func NewTaskDuration(dur time.Duration) *TaskDuration {
	return &TaskDuration{dur: dur, tMp: map[int64]time.Time{}}
}

func (t *TaskDuration) AddTask(gid int64) bool {
	now := time.Now()
	t.rw.Lock()
	defer t.rw.Unlock()
	last, ok := t.tMp[gid]
	if !ok || now.Sub(last) > t.dur {
		t.tMp[gid] = now
		return true
	}
	return false
}
