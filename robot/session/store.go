// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package session

import (
	"net"
	"sync"
	"time"

	"github.com/doublemo/balala/cores/proto"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

// Store session 存储
type Store struct {
	store  sync.Map
	logger log.Logger
}

// NewClient 创建一个新的session
// 默认情况session状态为匿名,等待通过认证后将修改sid,更改状态为已认证
func (ss *Store) NewClient(conn interface{}, sid string, readDeadline, writeDeadline time.Duration, maxMessageSize int64) *Client {
	var s Client
	s.id = sid
	s.flag = 0
	s.die = make(chan struct{})
	s.recvExitChan = make(chan struct{})
	s.sendExitChan = make(chan struct{})
	s.recvChan = make(chan []byte)
	s.sendChan = make(chan []byte, 1024)
	s.cacheBytes = make([]byte, 65535)
	s.readyedChan = make(chan struct{}, 2)
	s.protoTypes = proto.None
	s.logger = ss.logger
	s.params.Store(make(map[string]interface{}))

	switch c := conn.(type) {
	case *websocket.Conn:
		s.protoTypes = proto.Websocket
		s.websocketConn = c
		s.SetParam("RemoteAddr", c.RemoteAddr().String())

	case net.Conn:
		s.protoTypes = proto.Socket
		s.socketConn = c
		s.SetParam("RemoteAddr", c.RemoteAddr().String())
	}

	if s.id == "" {
		s.id = uuid.NewV4().String()
	}

	s.SetParam("CreateAt", time.Now())

	go s.recv(readDeadline, maxMessageSize)
	go s.send(writeDeadline)

	for i := 0; i < 2; i++ {
		<-s.readyedChan
	}

	close(s.readyedChan)
	ss.store.Store(s.id, &s)
	return &s
}

// Get 获取session
func (ss *Store) Get(sid string) *Client {
	s, ok := ss.store.Load(sid)
	if !ok {
		return nil
	}

	return s.(*Client)
}

// Remove 删除session
func (ss *Store) Remove(sid string) {
	ss.store.Delete(sid)
}

// Store 保存session
func (ss *Store) Store(s *Client) {
	ss.store.Store(s.id, s)
}

// RemoveAndExit 删除session并退出
func (ss *Store) RemoveAndExit(sid string) {
	if sess := ss.Get(sid); sess != nil {
		sess.Kicked()
	}

	ss.Remove(sid)
}

// NewStore 创建session存储器
func NewStore(logger log.Logger) *Store {
	return &Store{logger: logger}
}
