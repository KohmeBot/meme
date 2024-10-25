package meme

import (
	"github.com/wdvxdr1123/ZeroBot/message"
	"sync"
	"time"
)

type HelpTasks struct {
	rw    sync.RWMutex
	tasks map[int64]bool
}

func NewTasks() *HelpTasks {
	return &HelpTasks{tasks: make(map[int64]bool)}
}

func (t *HelpTasks) AddTask(uid int64) bool {
	t.rw.Lock()
	defer t.rw.Unlock()
	b, ok := t.tasks[uid]
	if !ok || !b {
		t.tasks[uid] = true
	}
	return !b
}

func (t *HelpTasks) Done(uid int64) {
	t.rw.Lock()
	defer t.rw.Unlock()
	t.tasks[uid] = false
}

type TaskDuration struct {
	dur    time.Duration
	rw     sync.RWMutex
	tMp    map[int64]time.Time
	lastId map[int64]message.MessageID
}

func NewTaskDuration(dur time.Duration) *TaskDuration {
	return &TaskDuration{dur: dur, tMp: map[int64]time.Time{}, lastId: make(map[int64]message.MessageID)}
}

func (t *TaskDuration) AddTask(gid int64) (bool, message.MessageID) {
	now := time.Now()
	t.rw.Lock()
	defer t.rw.Unlock()
	last, ok := t.tMp[gid]
	if !ok || now.Sub(last) > t.dur {
		t.tMp[gid] = now
		return true, message.MessageID{}
	}
	return false, t.lastId[gid]
}

func (t *TaskDuration) Done(gid int64, mid message.MessageID) {
	t.rw.Lock()
	defer t.rw.Unlock()
	t.lastId[gid] = mid
}
