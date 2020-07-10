// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package session

import (
	"sync"

	"github.com/doublemo/balala/cores/types"
)

// Store session 存储
type Store struct {
	store sync.Map
}

// NewClient 创建一个新的session
func (ss *Store) NewClient(id types.UID) *Client {
	var s Client
	s.id = id
	ss.store.Store(s.id, &s)
	return &s
}

// Get 获取session
func (ss *Store) Get(id types.UID) *Client {
	s, ok := ss.store.Load(id)
	if !ok {
		return nil
	}

	return s.(*Client)
}

// Remove 删除session
func (ss *Store) Remove(id types.UID) {
	ss.store.Delete(id)
}

// Store 保存session
func (ss *Store) Store(s *Client) {
	ss.store.Store(s.id, s)
}

// NewStore 创建session存储器
func NewStore() *Store {
	return &Store{}
}
