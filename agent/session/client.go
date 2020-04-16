// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package session

import (
	"crypto/rc4"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/doublemo/balala/cores/proto"
	"github.com/go-kit/kit/log"
	kitlog "github.com/go-kit/kit/log/level"
	"github.com/gorilla/websocket"
)

// 定义session状态
const (
	// FlagKeyexcg 是否已经交换完毕KEY
	FlagKeyexcg = 0x1

	// FlagEncrypt 是否可以开始加密
	FlagEncrypt = 0x2

	// FlagKickedOut 需要踢掉的用户
	FlagKickedOut = 0x4

	// FlagAuthorized 已授权
	FlagAuthorized = 0x8
)

// Client 连接信息
type Client struct {
	// id 唯一
	id string

	// protoTypes 客户端使用的通信协议
	// 目前支持 SOCKET, WEBSOCKET, GRPC 默认情况下SOCKET
	protoTypes proto.Types

	// websocketConn 客户端websocket通信支持
	// 绑住客户端webscoket连接状态
	websocketConn *websocket.Conn

	// socketConn 客户端TCP 通信支持
	socketConn net.Conn

	// flag 会话标记
	flag int32

	// encoder  数据加密
	encoder *rc4.Cipher

	// decoder 数据解密
	decoder *rc4.Cipher

	// recvChan 数据接入通道
	recvChan chan []byte

	// sendChan  数据发关通道
	sendChan chan []byte

	// recvExitChan  接收退出信息号
	recvExitChan chan struct{}

	// sendExitChan  输出退出信息号
	sendExitChan chan struct{}

	// readyedChan 准备就绪信号
	readyedChan chan struct{}

	// die 死亡信号
	die chan struct{}

	// cacheBytes 缓存空间
	cacheBytes []byte

	// logger 日志
	logger log.Logger

	// params 参数
	params atomic.Value

	// lock
	mutex sync.Mutex
}

func (s *Client) recv(readDeadline time.Duration, maxMessageSize int64) {
	defer func() {
		close(s.recvExitChan)
	}()

	s.readyedChan <- struct{}{}
	switch s.protoTypes {
	case proto.Socket:
		s.recvFromSocket(readDeadline)

	case proto.Websocket:
		s.recvFromWebSocket(readDeadline, maxMessageSize)

	case proto.None:
	}
}

func (s *Client) recvFromSocket(readDeadline time.Duration) {
	var logger log.Logger
	{
		logger = s.logger
	}

	header := make([]byte, 2)
	for {
		// 写入超时与读取超时
		s.socketConn.SetReadDeadline(time.Now().Add(readDeadline))
		n, err := io.ReadFull(s.socketConn, header)
		if err != nil {
			return
		}

		size := binary.BigEndian.Uint16(header)
		payload := make([]byte, size)
		n, err = io.ReadFull(s.socketConn, payload)
		if err != nil {
			kitlog.Error(logger).Log("error", "read payload failed", "reason", err.Error(), "size", n)
			return
		}

		select {
		case s.recvChan <- payload:
		case <-s.die:
			return

		case <-s.sendExitChan:
			return
		}

		if s.Flag()&FlagKickedOut != 0 {
			return
		}
	}
}

func (s *Client) recvFromWebSocket(readDeadline time.Duration, maxMessageSize int64) {
	var logger log.Logger
	{
		logger = s.logger
	}

	for {
		// 写入超时与读取超时
		s.websocketConn.SetReadLimit(maxMessageSize)
		s.websocketConn.SetReadDeadline(time.Now().Add(readDeadline))
		s.websocketConn.SetPongHandler(func(string) error {
			s.websocketConn.SetReadDeadline(time.Now().Add(readDeadline))
			return nil
		})

		frameType, payload, err := s.websocketConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				kitlog.Error(logger).Log("error", err)
			}
			return
		}

		if frameType != websocket.BinaryMessage {
			return
		}

		select {
		case s.recvChan <- payload[2:]:
		case <-s.die:
			return

		case <-s.sendExitChan:
			return
		}

		if s.Flag()&FlagKickedOut != 0 {
			return
		}
	}
}

func (s *Client) send(writeDeadline time.Duration) {
	var logger log.Logger
	{
		logger = s.logger
	}

	ticker := time.NewTicker(time.Second * 1)
	defer func() {
		ticker.Stop()
		close(s.sendExitChan)
	}()

	s.readyedChan <- struct{}{}
	for {
		select {
		case frame, ok := <-s.sendChan:
			if !ok {
				return
			}

			flag := s.Flag()
			if flag&FlagEncrypt != 0 {
				s.encoder.XORKeyStream(frame, frame)
			} else if flag&FlagKeyexcg != 0 {
				flag &^= FlagKeyexcg
				flag |= FlagEncrypt
				s.Flag(flag)
			}

			if err := s.write(frame, writeDeadline); err != nil {
				kitlog.Error(logger).Log("error", err)
				return
			}

		case <-ticker.C:
			// websocket ping
			if s.protoTypes != proto.Websocket {
				continue
			}

			s.websocketConn.SetWriteDeadline(time.Now().Add(writeDeadline))
			if err := s.websocketConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-s.die:
			return

		case <-s.recvExitChan:
			return
		}

		if s.Flag()&FlagKickedOut != 0 {
			return
		}
	}
}

func (s *Client) write(frame []byte, writeDeadline time.Duration) (err error) {
	switch s.protoTypes {
	case proto.Socket:

		if writeDeadline.Nanoseconds() > 0 {
			s.socketConn.SetWriteDeadline(time.Now().Add(writeDeadline))
		}

		err = s.writeToSocket(frame)
		if err != nil {
			return
		}

	case proto.Websocket:

		if writeDeadline.Nanoseconds() > 0 {
			s.websocketConn.SetWriteDeadline(time.Now().Add(writeDeadline))
		}

		err = s.writeToWebSocket(frame)
		if err != nil {
			return
		}

	case proto.None:
	}

	return
}

func (s *Client) writeToSocket(frame []byte) error {
	size := len(frame)
	binary.BigEndian.PutUint16(s.cacheBytes, uint16(size))
	copy(s.cacheBytes[2:], frame)
	_, err := s.socketConn.Write(s.cacheBytes[:size+2])
	return err
}

func (s *Client) writeToWebSocket(frame []byte) (err error) {
	size := len(frame)
	binary.BigEndian.PutUint16(s.cacheBytes, uint16(size))
	copy(s.cacheBytes[2:], frame)

	var w io.WriteCloser
	{
		w, err = s.websocketConn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			return
		}
	}

	w.Write(s.cacheBytes[:size+2])
	if err = w.Close(); err != nil {
		return
	}

	return err
}

// Flag 客户端状态
func (s *Client) Flag(args ...int32) int32 {
	if len(args) > 0 {
		atomic.StoreInt32(&s.flag, args[0])
		return args[0]
	}

	return atomic.LoadInt32(&s.flag)
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

// Param 获取session数据
func (s *Client) Param(key string) (interface{}, bool) {
	m := s.params.Load().(map[string]interface{})
	v, ok := m[key]
	return v, ok
}

// SetLogger 设置日志处理
func (s *Client) SetLogger(logger log.Logger) {
	s.logger = logger
}

// GetLogger 获取日志处理
func (s *Client) GetLogger() log.Logger {
	return s.logger
}

// SetEncoder RC4加密
func (s *Client) SetEncoder(encoder *rc4.Cipher) {
	s.encoder = encoder
}

// SetDecoder rc4解密
func (s *Client) SetDecoder(decoder *rc4.Cipher) {
	s.decoder = decoder
}

// GetRecvChan 获取接收通道
func (s *Client) GetRecvChan() chan []byte {
	return s.recvChan
}

// GetRecvExitChan 获取退出通道
func (s *Client) GetRecvExitChan() <-chan struct{} {
	return s.recvExitChan
}

// GetSendExitChan 信息推送退出通道
func (s *Client) GetSendExitChan() <-chan struct{} {
	return s.sendExitChan
}

// Kicked 踢掉客户端
func (s *Client) Kicked() {
	flag := atomic.LoadInt32(&s.flag)
	if flag&FlagKickedOut != 0 {
		return
	}

	flag |= FlagKickedOut
	atomic.StoreInt32(&s.flag, flag)
	close(s.die)
}

// ID 获取ID
func (s *Client) ID() string {
	return s.id
}
