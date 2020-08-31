package session

import (
	"sync"
	"sync/atomic"
)

type Param struct {
	Key   string
	Value string
}

// Client 连接信息
type Client struct {

	// id 连接唯一id
	id string

	// params 参数
	params atomic.Value

	// lock
	mutex sync.Mutex
}

// ID ...
func (s Client) ID() string {
	return s.id
}

// SetParam 设置session数据
func (s *Client) SetParam(params ...*Param) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	m1 := s.params.Load().(map[string]string)
	m2 := make(map[string]string)
	for k, v := range m1 {
		m2[k] = v
	}

	for _, v := range params {
		m2[v.Key] = v.Value
	}
	s.params.Store(m2)
}

// RemoveParam 设置session数据
func (s *Client) RemoveParam(keys ...string) {
	mkeys := make(map[string]bool)
	for _, k := range keys {
		mkeys[k] = true
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	m1 := s.params.Load().(map[string]string)
	m2 := make(map[string]string)
	for k, v := range m1 {
		if mkeys[k] {
			continue
		}
		m2[k] = v
	}
	s.params.Store(m2)
}

// Param 获取session数据
func (s *Client) Param(key string) (string, bool) {
	m := s.params.Load().(map[string]string)
	v, ok := m[key]
	return v, ok
}
