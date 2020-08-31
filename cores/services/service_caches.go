// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package services

import (
	"encoding/json"
	"math/rand"
	"sync"

	"github.com/doublemo/balala/cores/types"
)

// Caches 服务信息绑存
type Caches struct {
	records    map[int32][]*Options
	roundRobin map[int32]int
	mutex      sync.RWMutex
}

// Reset 重置
func (caches *Caches) Reset() {
	caches.mutex.Lock()
	defer caches.mutex.Unlock()
	caches.records = make(map[int32][]*Options)
}

// Store 存储来Options对象
func (caches *Caches) Store(o *Options) {
	if o.ID < 1 {
		return
	}

	caches.mutex.Lock()
	defer caches.mutex.Unlock()

	if _, ok := caches.records[o.ID]; !ok {
		caches.records[o.ID] = []*Options{o}
		return
	}

	caches.records[o.ID] = append(caches.records[o.ID], o)
}

// StoreFromString 存储来Options对象
func (caches *Caches) StoreFromString(s string) (*Options, error) {
	o := Options{}
	if err := json.Unmarshal([]byte(s), &o); err != nil {
		return nil, err
	}

	caches.Store(&o)
	return &o, nil
}

// RndOnce 随机一个
func (caches *Caches) RndOnce(id int32) (*Options, bool) {
	caches.mutex.RLock()
	defer caches.mutex.RUnlock()
	m, ok := caches.records[id]
	if !ok || len(m) < 1 {
		return nil, false
	}

	if len(m) == 1 {
		if m[0].Priority > 0 {
			return m[0], true
		}

		return nil, false
	}

	sumNumber := 0
	for _, i := range m {
		sumNumber += i.Priority
	}

	if sumNumber < 1 {
		return nil, false
	}

	rndArray := make([]int, 0)
	for idx, i := range m {
		if i.Priority < 1 {
			continue
		}

		mod := int(types.Round(float64(i.Priority)/float64(sumNumber), 2) * 100)
		for i := 0; i < mod; i++ {
			rndArray = append(rndArray, idx)
		}
	}

	if len(rndArray) < 1 {
		return nil, false
	}

	for i := range rndArray {
		j := rand.Intn(i + 1)
		rndArray[i], rndArray[j] = rndArray[j], rndArray[i]
	}

	index := rndArray[rand.Intn(len(rndArray))]
	return m[index], true
}

// RoundRobinOnce 循环一个
func (caches *Caches) RoundRobinOnce(id int32) (*Options, bool) {
	caches.mutex.RLock()
	defer caches.mutex.RUnlock()

	m, ok := caches.records[id]
	if !ok || len(m) < 1 {
		return nil, false
	}

	if len(m) == 1 {
		if m[0].Priority > 0 {
			return m[0], true
		}

		return nil, false
	}

	rridx := caches.roundRobin[id]
	if rridx >= len(m) {
		rridx = 0
	}

	var mr *Options
	old := rridx
	for {
		mrr := m[rridx]
		if mrr.Priority > 0 {
			mr = mrr
			break
		}

		rridx++
		if rridx >= len(m) {
			rridx = 0
		}

		if rridx == old {
			mr = nil
			break
		}
	}

	if rridx+1 >= len(m) {
		caches.roundRobin[id] = 0
	} else {
		caches.roundRobin[id] = rridx + 1
	}

	return mr, mr != nil
}

// NewCaches 服务信息缓存
func NewCaches() *Caches {
	return &Caches{
		records:    make(map[int32][]*Options),
		roundRobin: make(map[int32]int),
	}
}
