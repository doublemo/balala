package session

import (
	"sync"
	"sync/atomic"

	"github.com/doublemo/balala/cores/types"
)

// Client 连接信息
type Client struct {

	// id 连接唯一id
	id types.UID

	// params 参数
	params atomic.Value

	// lock
	mutex sync.Mutex
}

// ID ...
func (s Client) ID() types.UID {
	return s.id
}

// SetParam 设置session数据
func (s *Client) SetParam(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	m1 := s.params.Load().(map[string]interface{})
	m2 := make(map[string]interface{})
	for k, v := range m1 {
		m2[k] = v
	}

	m2[key] = value
	s.params.Store(m2)
}

// RemoveParam 设置session数据
func (s *Client) RemoveParam(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	m1 := s.params.Load().(map[string]interface{})
	m2 := make(map[string]interface{})
	for k, v := range m1 {
		if k == key {
			continue
		}

		m2[k] = v
	}
	s.params.Store(m2)
}

// Param 获取session数据
func (s *Client) Param(key string) (interface{}, bool) {
	m := s.params.Load().(map[string]interface{})
	v, ok := m[key]
	return v, ok
}
