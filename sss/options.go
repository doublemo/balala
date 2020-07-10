// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package sss

import (
	"errors"
	"os"
	"sync/atomic"
	"unsafe"

	"github.com/doublemo/balala/cores/alias"
	"github.com/doublemo/balala/cores/networks"
)

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
	DialTimeout int `alias:"dialtimeout" default:"3"`

	// DialKeepAlive If DialKeepAlive is 0, it defaults to 3s
	DialKeepAlive int `alias:"dialkeepalive" default:"3"`
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

// TracerOptions 请求运行追踪
type TracerOptions struct {
	// ReporterURL 追踪服务地址 eg:http://192.168.31.20:9411/api/v2/spans
	ReporterURL string `alias:"reporterurl"`
}

// Clone ETCDOptions
func (o *TracerOptions) Clone() *TracerOptions {
	return &TracerOptions{
		ReporterURL: o.ReporterURL,
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

	// Domain string 提供服务的域名
	Domain string `alias:"domain"`

	// Priority 优先级
	Priority int `alias:"priority" default:"1"`

	// GRPC 将支持GRPC服务
	GRPC *GRPCOptions `alias:"grpc"`

	// ETCD etcd
	ETCD *ETCDOptions `alias:"etcd"`

	// ServiceSecurityKey JWT 服务之通信认证
	ServiceSecurityKey string `alias:"servicesecuritykey"`

	// Tracer 请求运行追踪
	Tracer *TracerOptions `alias:"tracer"`
}

// Clone 克隆配置文件防止调用配置文件时造成冲突
func (o *Options) Clone() *Options {
	copy := Options{}
	copy.ID = o.ID
	copy.Runmode = o.Runmode
	copy.Priority = o.Priority
	if o.LocalIP == "" {
		if m, err := networks.LocalIP(); err == nil {
			copy.LocalIP = m.String()
		}
	} else {
		copy.LocalIP = o.LocalIP
	}

	if o.Domain == "" {
		copy.Domain = copy.LocalIP
	} else {
		copy.Domain = o.Domain
	}

	if o.GRPC != nil {
		copy.GRPC = o.GRPC.Clone()
	}

	if o.ETCD != nil {
		copy.ETCD = o.ETCD.Clone()
	}

	if o.Tracer != nil {
		copy.Tracer = o.Tracer.Clone()
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
