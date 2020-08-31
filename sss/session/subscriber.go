// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package session

import (
	"errors"
	"sync"

	"github.com/doublemo/balala/sss/proto/pb"
)

var (
	// ErrAlreadyExists 存储信息已经存在
	ErrAlreadyExists = errors.New("ErrAlreadyExists")

	// ErrNotFound 存储信息不存在
	ErrNotFound = errors.New("ErrNotFound")
)

// Subscriber 订阅人
type Subscriber struct {
	id        string
	serviceID int32
	recvchan  chan *pb.SessionStateServerAPI_BroadcastResponse
	events    map[int32]bool
}

func (subscriber *Subscriber) GetID() string {
	return subscriber.id
}

func (subscriber *Subscriber) GetServiceID() int32 {
	return subscriber.serviceID
}

func (subscriber *Subscriber) GetRecv() chan *pb.SessionStateServerAPI_BroadcastResponse {
	return subscriber.recvchan
}

// NewSubscriber 创建订阅者
func NewSubscriber(id string, serviceID int32, events ...int32) *Subscriber {
	subscriber := &Subscriber{
		id:        id,
		serviceID: serviceID,
		recvchan:  make(chan *pb.SessionStateServerAPI_BroadcastResponse, 128),
		events:    make(map[int32]bool),
	}

	for _, event := range events {
		subscriber.events[event] = true
	}

	return subscriber
}

// SubscribeStore 订阅管理
type SubscribeStore struct {
	stores    map[int32][]*Subscriber
	storesMap map[string]bool
	mutex     sync.RWMutex
}

// NewSubscriber 创建新订阅信息
func (subscribeStore *SubscribeStore) NewSubscriber(id string, serviceID int32, events []int32) (*Subscriber, error) {
	subscribeStore.mutex.Lock()
	defer subscribeStore.mutex.Unlock()

	if subscribeStore.storesMap[id] {
		return nil, ErrAlreadyExists
	}

	subscribeStore.storesMap[id] = true
	if _, ok := subscribeStore.stores[serviceID]; ok {
		subscribeStore.stores[serviceID] = make([]*Subscriber, 0)
	}

	subscriber := NewSubscriber(id, serviceID, events...)
	subscribeStore.stores[serviceID] = append(subscribeStore.stores[serviceID], subscriber)
	return subscriber, nil
}

// Store 存储订阅信息
func (subscribeStore *SubscribeStore) Store(subscriber *Subscriber) error {
	if len(subscriber.GetID()) < 1 || subscriber.GetServiceID() < 1 {
		return errors.New("the subscription is invalid")
	}

	subscribeStore.mutex.Lock()
	defer subscribeStore.mutex.Unlock()

	if subscribeStore.storesMap[subscriber.GetID()] {
		return ErrAlreadyExists
	}

	subscribeStore.storesMap[subscriber.GetID()] = true
	if _, ok := subscribeStore.stores[subscriber.GetServiceID()]; ok {
		subscribeStore.stores[subscriber.GetServiceID()] = make([]*Subscriber, 0)
	}

	subscribeStore.stores[subscriber.GetServiceID()] = append(subscribeStore.stores[subscriber.GetServiceID()], subscriber)
	return nil
}

// Remove 删除订阅
func (subscribeStore *SubscribeStore) Remove(id string) {
	subscribeStore.mutex.Lock()
	if !subscribeStore.storesMap[id] {
		subscribeStore.mutex.Unlock()
		return
	}

	var mserviceID int32
	for serviceID, subscribers := range subscribeStore.stores {
		subscribeStore.mutex.Unlock()
		for _, subscriber := range subscribers {
			if subscriber.GetID() == id {
				mserviceID = serviceID
				break
			}
		}
		subscribeStore.mutex.Lock()
		if mserviceID > 0 {
			break
		}
	}

	if mserviceID < 1 {
		subscribeStore.mutex.Unlock()
		return
	}

	subscribers := subscribeStore.stores[mserviceID]
	if subscribers == nil {
		subscribeStore.mutex.Unlock()
		return
	}

	newSubscribers := make([]*Subscriber, len(subscribers)-1)
	idx := 0
	for _, s := range subscribers {
		if s.GetID() == id {
			continue
		}
		newSubscribers[idx] = s
		idx++
	}

	delete(subscribeStore.storesMap, id)
	subscribeStore.stores[mserviceID] = newSubscribers
	subscribeStore.mutex.Unlock()
}

// RemoveByServiceIDAndID 删除订阅
func (subscribeStore *SubscribeStore) RemoveByServiceIDAndID(id string, serviceID int32) {
	if len(id) < 1 || serviceID < 1 {
		return
	}

	subscribeStore.mutex.Lock()
	defer subscribeStore.mutex.Unlock()

	if !subscribeStore.storesMap[id] {
		return
	}

	subscribers := subscribeStore.stores[serviceID]
	if subscribers == nil {
		return
	}

	newSubscribers := make([]*Subscriber, len(subscribers)-1)
	idx := 0
	for _, s := range subscribers {
		if s.GetID() == id {
			continue
		}
		newSubscribers[idx] = s
		idx++
	}

	delete(subscribeStore.storesMap, id)
	subscribeStore.stores[serviceID] = newSubscribers
}

// GetSubscriber 获取订阅
func (subscribeStore *SubscribeStore) GetSubscriber(id string) (*Subscriber, error) {
	subscribeStore.mutex.RLock()
	if !subscribeStore.storesMap[id] {
		subscribeStore.mutex.RUnlock()
		return nil, ErrNotFound
	}

	for _, subscribers := range subscribeStore.stores {
		subscribeStore.mutex.RUnlock()
		for _, subscriber := range subscribers {
			if subscriber.GetID() == id {
				subscribeStore.mutex.RUnlock()
				return subscriber, nil
			}
		}
		subscribeStore.mutex.RLock()
	}

	subscribeStore.mutex.RUnlock()
	return nil, ErrNotFound
}

// GetSubscriberByServiceIDAndID 获取订阅
func (subscribeStore *SubscribeStore) GetSubscriberByServiceIDAndID(id string, serviceID int32) (*Subscriber, error) {
	subscribeStore.mutex.RLock()
	defer subscribeStore.mutex.RUnlock()

	if !subscribeStore.storesMap[id] {
		return nil, ErrNotFound
	}

	subscribers := subscribeStore.stores[serviceID]
	if subscribers == nil {
		return nil, ErrNotFound
	}

	for _, s := range subscribers {
		if s.GetID() == id {
			return s, nil
		}
	}

	return nil, ErrNotFound
}

// GetSubscribersByServiceID 获取订阅
func (subscribeStore *SubscribeStore) GetSubscribersByServiceID(serviceID int32) []*Subscriber {
	subscribeStore.mutex.RLock()
	defer subscribeStore.mutex.RUnlock()

	subscribers := subscribeStore.stores[serviceID]
	if subscribers == nil {
		return make([]*Subscriber, 0)
	}

	newSubscribers := make([]*Subscriber, len(subscribers))
	copy(newSubscribers[0:], subscribers[0:])
	return newSubscribers
}

// GetSubscribers 获取订阅
func (subscribeStore *SubscribeStore) GetSubscribers() []*Subscriber {
	subscribeStore.mutex.RLock()
	defer subscribeStore.mutex.RUnlock()

	newSubscribers := make([]*Subscriber, len(subscribeStore.storesMap))
	start := 0
	for _, subscribers := range subscribeStore.stores {
		copy(newSubscribers[start:start+len(subscribers)], subscribers[0:])
		start += len(subscribers)
	}
	return newSubscribers
}

// NewSubscribeStore create
func NewSubscribeStore() *SubscribeStore {
	return &SubscribeStore{
		stores:    make(map[int32][]*Subscriber),
		storesMap: make(map[string]bool),
	}
}
