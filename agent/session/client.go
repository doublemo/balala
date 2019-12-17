// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package session

import (
	"crypto/rc4"
	"encoding/binary"
	"errors"
	"io"
	"net"
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

// Callback 回调
type Callback func(*Client, []byte) ([]byte, error)

// Client 客户
type Client struct {
	// Proto 客户端使用的通信协议
	// 目前支持 SOCKET, WEBSOCKET, GRPC 默认情况下SOCKET
	Proto proto.Types

	// WebsocketConn 客户端websocket通信支持
	// 绑住客户端webscoket连接状态
	WebsocketConn *websocket.Conn

	// SocketConn 客户端TCP 通信支持
	SocketConn net.Conn

	// userAgent  客户信息
	UserAgent string

	// CreateAt session创建时间
	CreateAt time.Time

	// PacketCounter 包数量统计,也用验证客户端发来是否重复和客户端统计保持一致
	PacketCounter int

	// Sid sessionid
	Sid string

	// UserID user id
	UserID uint64

	// Device 玩家设备
	Device int8

	// remoteAddr 客户端地址
	remoteAddr string

	// encoder  数据加密
	encoder *rc4.Cipher

	// decoder 数据解密
	decoder *rc4.Cipher

	// flag 会话标记
	flag int32

	// recvChan 数据接入通道
	recvChan chan []byte

	// sendChan  数据发关通道
	sendChan chan []byte

	// exitChan  退出信息号
	exitChan chan struct{}

	// recvExitChan  接收退出信息号
	recvExitChan chan struct{}

	// sendExitChan  输出退出信息号
	sendExitChan chan struct{}

	// cacheBytes 缓存空间
	cacheBytes []byte

	// logger 日志
	logger log.Logger
}

// Send 向客户端推送信息
func (s *Client) Send(frame []byte) error {
	if frame == nil {
		return nil
	}
	select {
	case s.sendChan <- frame:
	default:
		kitlog.Warn(s.logger).Log("error", "chanfull", "sid", s.Sid)
		return errors.New("chanfull")
	}

	return nil
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

// Flag 客户端状态
func (s *Client) Flag(args ...int32) int32 {
	if len(args) > 0 {
		atomic.StoreInt32(&s.flag, args[0])
		return args[0]
	}

	return atomic.LoadInt32(&s.flag)
}

// Call 加密
func (s *Client) Call(frame []byte, route Callback) ([]byte, error) {
	// 实现数据解密
	if s.Flag()&FlagEncrypt != 0 {
		s.decoder.XORKeyStream(frame, frame)
	}

	return route(s, frame)
}

// Kicked 踢掉客户端
func (s *Client) Kicked() {
	flag := atomic.LoadInt32(&s.flag)
	if flag&FlagKickedOut != 0 {
		return
	}

	flag |= FlagKickedOut
	atomic.StoreInt32(&s.flag, flag)
	close(s.exitChan)
}

func (s *Client) recv(readDeadline time.Duration, maxMessageSize int64, readyed chan bool) {
	defer func() {
		close(s.recvExitChan)
	}()

	readyed <- true
	switch s.Proto {
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
		s.SocketConn.SetReadDeadline(time.Now().Add(readDeadline))
		n, err := io.ReadFull(s.SocketConn, header)
		if err != nil {
			return
		}

		size := binary.BigEndian.Uint16(header)
		payload := make([]byte, size)
		n, err = io.ReadFull(s.SocketConn, payload)

		if err != nil {
			kitlog.Error(logger).Log("error", "read payload failed", "reason", err.Error(), "size", n)
			return
		}

		select {
		case s.recvChan <- payload:
		case <-s.exitChan:
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
		s.WebsocketConn.SetReadLimit(maxMessageSize)
		s.WebsocketConn.SetReadDeadline(time.Now().Add(readDeadline))
		s.WebsocketConn.SetPongHandler(func(string) error {
			s.WebsocketConn.SetReadDeadline(time.Now().Add(readDeadline))
			return nil
		})

		frameType, payload, err := s.WebsocketConn.ReadMessage()
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
		case <-s.exitChan:
			return
		case <-s.sendExitChan:
			return
		}

		if s.Flag()&FlagKickedOut != 0 {
			return
		}
	}
}

func (s *Client) send(writeDeadline time.Duration, readyed chan bool) {
	var logger log.Logger
	{
		logger = s.logger
	}

	ticker := time.NewTicker(time.Second * 1)
	defer func() {
		ticker.Stop()
		close(s.sendExitChan)
	}()

	readyed <- true
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
			if s.Proto != proto.Websocket {
				continue
			}

			s.WebsocketConn.SetWriteDeadline(time.Now().Add(writeDeadline))
			if err := s.WebsocketConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-s.exitChan:
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
	switch s.Proto {
	case proto.Socket:

		if writeDeadline.Nanoseconds() > 0 {
			s.SocketConn.SetWriteDeadline(time.Now().Add(writeDeadline))
		}

		err = s.writeToSocket(frame)
		if err != nil {
			return
		}

	case proto.Websocket:

		if writeDeadline.Nanoseconds() > 0 {
			s.WebsocketConn.SetWriteDeadline(time.Now().Add(writeDeadline))
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
	_, err := s.SocketConn.Write(s.cacheBytes[:size+2])
	return err
}

func (s *Client) writeToWebSocket(frame []byte) (err error) {
	size := len(frame)
	binary.BigEndian.PutUint16(s.cacheBytes, uint16(size))
	copy(s.cacheBytes[2:], frame)

	var w io.WriteCloser
	{
		w, err = s.WebsocketConn.NextWriter(websocket.BinaryMessage)
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
