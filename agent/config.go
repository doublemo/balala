// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package agent 代理服务器
package agent

import (
	"errors"
	"os"
	"sync/atomic"
	"unsafe"

	"github.com/doublemo/balala/cores/alias"
	"github.com/doublemo/balala/cores/networks"
)

// HTTPOptions http配置
type HTTPOptions struct {
	// Addr 监听地址
	Addr string `alias:"addr" default:":9090"`

	// ReadTimeout http服务读取超时
	ReadTimeout int `alias:"readtimeout" default:"10"`

	// WriteTimeout http服务写入超时
	WriteTimeout int `alias:"writetimeout" default:"10"`

	// MaxHeaderBytes  http内容大小限制
	MaxHeaderBytes int `alias:"maxheaderbytes" default:"1048576"`

	// SSL ssl 支持
	SSL bool `alias:"ssl" default:"false"`

	// Key 证书key
	Key string `alias:"key"`

	// SSLCert 证书
	Cert string `alias:"cert"`
}

// Clone 克隆HTTPOptions
func (o *HTTPOptions) Clone() *HTTPOptions {
	return &HTTPOptions{
		Addr:           o.Addr,
		ReadTimeout:    o.ReadTimeout,
		WriteTimeout:   o.WriteTimeout,
		MaxHeaderBytes: o.MaxHeaderBytes,
		SSL:            o.SSL,
		Key:            o.Key,
		Cert:           o.Cert,
	}
}

// SocketOptions tcp参数
type SocketOptions struct {
	// Addr 监听地址
	Addr string `alias:"addr" default:":9091"`

	// ReadBufferSize 读取缓存大小 32767
	ReadBufferSize int `alias:"readbuffersize" default:"32767"`

	// WriteBufferSize 写入缓存大小 32767
	WriteBufferSize int `alias:"writebuffersize" default:"32767"`

	// ReadDeadline 读取超时
	ReadDeadline int `alias:"readdeadline" default:"310"`

	// WriteDeadline 写入超时
	WriteDeadline int `alias:"writedeadline"`
}

// Clone SocketOptions
func (o *SocketOptions) Clone() *SocketOptions {
	return &SocketOptions{
		Addr:            o.Addr,
		ReadBufferSize:  o.ReadBufferSize,
		WriteBufferSize: o.WriteBufferSize,
		ReadDeadline:    o.ReadDeadline,
		WriteDeadline:   o.WriteDeadline,
	}
}

// GRPCOptions GRP参数
type GRPCOptions struct {
	// Addr 监听地址
	Addr string `alias:"addr" default:":9092"`
}

// Clone GRPCOptions
func (o *GRPCOptions) Clone() *GRPCOptions {
	return &GRPCOptions{
		Addr: o.Addr,
	}
}

// WebSocketOptions websocket配置
type WebSocketOptions struct {
	// Addr 监听地址
	Addr string `alias:"addr" default:":9093"`

	// ReadBufferSize 读取缓存大小 32767
	ReadBufferSize int `alias:"readbuffersize" default:"32767"`

	// WriteBufferSize 写入缓存大小 32767
	WriteBufferSize int `alias:"writebuffersize" default:"32767"`

	// MaxMessageSize WebSocket每帧最在数据大小
	MaxMessageSize int64 `alias:"maxmessagesize" default:"1024"`

	// ReadDeadline 读取超时
	ReadDeadline int `alias:"readdeadline" default:"310"`

	// WriteDeadline 写入超时
	WriteDeadline int `alias:"writedeadline"`
}

// Clone WebSocketOptions
func (o *WebSocketOptions) Clone() *WebSocketOptions {
	return &WebSocketOptions{
		Addr:            o.Addr,
		ReadBufferSize:  o.ReadBufferSize,
		WriteBufferSize: o.WriteBufferSize,
		MaxMessageSize:  o.MaxMessageSize,
		ReadDeadline:    o.ReadDeadline,
		WriteDeadline:   o.WriteDeadline,
	}
}

// ETCDOptions etcd参数
type ETCDOptions struct {
	// Address etcd 服务器地址
	Address []string `alias:"addr"`

	// Frefix etcd 存储值前缀
	Frefix string `alias:"frefix" default:"/services/balala"`

	// CACert etcd CA证书地址
	CACert string `alias:"cacert"`

	// Cert etcd 证书地址
	Cert string `alias:"cert"`

	// Key  etcd 证书key地址
	Key string `alias:"key"`

	// Username etcd 验证用户名
	Username string `alias:"username"`

	// Password etcd 验证密码
	Password string `alias:"password"`

	// DialTimeout if DialTimeout is 0, it defaults to 3s
	DialTimeout int `alias:"dialtimeout"`

	// DialKeepAlive If DialKeepAlive is 0, it defaults to 3s
	DialKeepAlive int `alias:"dialkeepalive"`
}

// Clone ETCDOptions
func (o *ETCDOptions) Clone() *ETCDOptions {
	return &ETCDOptions{
		Address:       o.Address,
		Frefix:        o.Frefix,
		CACert:        o.CACert,
		Cert:          o.Cert,
		Key:           o.Key,
		Username:      o.Username,
		Password:      o.Password,
		DialTimeout:   o.DialTimeout,
		DialKeepAlive: o.DialKeepAlive,
	}
}

// Options 配置参数
type Options struct {
	// 当前服务的唯一标识
	ID string `alias:"id" default:"agent"`

	// Runmode 运行模式
	Runmode string `alias:"runmode" default:"pord"`

	// LocalIP 当前服务器IP地址
	LocalIP string `alias:"localip"`

	// HTTP http(s) 监听端口
	// 利用http实现信息GET/POST, webscoket 也会这个端口甚而上实现
	HTTP *HTTPOptions `alias:"http"`

	// Socket 将支持tcp流服务
	Socket *SocketOptions `alias:"socket"`

	// GRPC 将支持GRPC服务
	GRPC *GRPCOptions `alias:"grpc"`

	// WebSocket 将支持WebSocket服务
	WebSocket *WebSocketOptions `alias:"websocket"`

	// ETCD etcd
	ETCD *ETCDOptions `alias:"etcd"`

	// ServiceSecurityKey JWT 服务之通信认证
	ServiceSecurityKey string `alias:"servicesecuritykey"`
}

// Clone 克隆配置文件防止调用配置文件时造成冲突
func (o *Options) Clone() *Options {
	copy := Options{}
	copy.ID = o.ID
	copy.Runmode = o.Runmode
	if o.LocalIP == "" {
		if m, err := networks.LocalIP(); err == nil {
			o.LocalIP = m.String()
		}
	} else {
		copy.LocalIP = o.LocalIP
	}

	if o.HTTP != nil {
		copy.HTTP = o.HTTP.Clone()
	}

	if o.Socket != nil {
		copy.Socket = o.Socket.Clone()
	}

	if o.GRPC != nil {
		copy.GRPC = o.GRPC.Clone()
	}

	if o.WebSocket != nil {
		copy.WebSocket = o.WebSocket.Clone()
	}

	if o.ETCD != nil {
		copy.ETCD = o.ETCD.Clone()
	}

	copy.ServiceSecurityKey = o.ServiceSecurityKey
	return &copy
}

// ConfigureOptions 配置文件服务
type ConfigureOptions struct {
	// 配置文件地址
	fp string

	// opts 配置信息
	opts unsafe.Pointer
}

// Read 加载配置文件
func (conf *ConfigureOptions) Read() *Options {
	return (*Options)(atomic.LoadPointer(&conf.opts)).Clone()
}

// Load 加载配置文件
func (conf *ConfigureOptions) Load() error {
	if conf.fp == "" {
		return errors.New("config file does not exist")
	}

	if _, err := os.Stat(conf.fp); os.IsNotExist(err) {
		return errors.New("config file does not exist")
	}

	opts := Options{}
	if err := alias.BindWithConfFile(conf.fp, &opts); err != nil {
		return err
	}

	conf.Reset(&opts)
	return nil
}

// Reset 重置配置文件
func (conf *ConfigureOptions) Reset(opts *Options) {
	atomic.StorePointer(&conf.opts, unsafe.Pointer(opts))
}

// NewConfigureOptions 创建配置文件
func NewConfigureOptions(fp string, opts *Options) *ConfigureOptions {
	c := &ConfigureOptions{fp: fp}
	if opts == nil {
		opts = &Options{}
	}
	atomic.StorePointer(&c.opts, unsafe.Pointer(opts))
	return c
}
