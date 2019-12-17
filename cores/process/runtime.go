// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package process

import (
	"sort"
	"sync"
	"sync/atomic"
)

// RuntimeActor 协程
type RuntimeActor struct {
	id int32

	stoped int32

	Exec func() error

	Interrupt func(error)

	Close func()
}

// Clone 复制协程
func (ra *RuntimeActor) Clone() *RuntimeActor {
	return &RuntimeActor{
		id:        ra.id,
		stoped:    ra.stoped,
		Exec:      ra.Exec,
		Interrupt: ra.Interrupt,
		Close:     ra.Close,
	}
}

type runtimeError struct {
	id  int32
	err error
}

// RuntimeContainer 协程管理盒子
type RuntimeContainer struct {
	actors  sync.Map
	counter int32
}

// Add 增加协程服务到盒子
func (rc *RuntimeContainer) Add(actor *RuntimeActor, status bool) int32 {
	if actor == nil {
		return 0
	}

	actor.id = atomic.AddInt32(&rc.counter, 1)
	if status {
		actor.stoped = 1
	} else {
		actor.stoped = 0
	}

	rc.actors.Store(actor.id, actor)
	return actor.id
}

// Run 运行盒子内
func (rc *RuntimeContainer) Run() error {
	errors := make(chan runtimeError)
	defer close(errors)

	actors := rc.sortRuntimeContainer(1)
	waitCounter := 0
	for _, actor := range actors {
		if atomic.LoadInt32(&actor.stoped) == 0 {
			continue
		}

		waitCounter++
		atomic.StoreInt32(&actor.stoped, 0)
		go func(a *RuntimeActor) {
			errors <- runtimeError{id: a.id, err: a.Exec()}
		}(actor)
	}

	if waitCounter < 1 {
		return nil
	}

	var (
		doneCounter int
		err         error
	)

	for e := range errors {
		doneCounter++
		if m, ok := rc.actors.Load(e.id); ok {
			actor := m.(*RuntimeActor)
			if atomic.LoadInt32(&actor.stoped) == 0 {
				actor.Interrupt(e.err)
				atomic.StoreInt32(&actor.stoped, 1)
			}
		}

		if err == nil && e.err != nil {
			err = e.err
		}

		if doneCounter >= waitCounter {
			break
		}
	}
	return err
}

// Stop 关闭盒子内所有服务
func (rc *RuntimeContainer) Stop() {
	actors := rc.sortRuntimeContainer(0)
	for i := len(actors) - 1; i >= 0; i-- {
		actor := (*RuntimeActor)(actors[i])
		if atomic.LoadInt32(&actor.stoped) == 1 {
			continue
		}

		actor.Close()
	}
}

func (rc *RuntimeContainer) sortRuntimeContainer(status int32) []*RuntimeActor {
	data := make([]*RuntimeActor, 0)
	rc.actors.Range(func(k, v interface{}) bool {
		actor := v.(*RuntimeActor)
		if atomic.LoadInt32(&actor.stoped) == status {
			data = append(data, actor)
		}

		return true
	})

	sort.Slice(data, func(a, b int) bool {
		actorA := (*RuntimeActor)(data[a])
		actorB := (*RuntimeActor)(data[b])
		if actorA.id > actorB.id {
			return false
		}
		return true
	})

	return data
}

// NewRuntimeContainer 创建盒子
func NewRuntimeContainer() *RuntimeContainer {
	c := &RuntimeContainer{}
	return c
}
