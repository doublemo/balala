// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package service 服务
package service

import (
	"encoding/json"
	"strconv"
)

// 定义软件信息
const (
	// 软件运行时显示的服务名称
	Name string = "agent"

	// 软件运行时用于识别的服务ID
	ID int32 = 1
)

// Value 服务存储的值
type Value struct {
	ID            int32  `json:"id"`
	Name          string `json:"n"`
	LocalID       string `json:"ip"`
	MachineID     string `json:"-"`
	GRPCAddr      string `json:"ga"`
	HTTPAddr      string `json:"ha"`
	SocketAddr    string `json:"sa"`
	WebsocketAddr string `json:"wsa"`
	Frefix        string `json:"-"`
}

// Key 服务KEY,用于ETCD服务发现
func (v *Value) Key() string {
	machineID := ""
	if v.MachineID != "" {
		machineID = "/" + v.MachineID
	}

	return v.Frefix + "/" + strconv.FormatInt(int64(v.ID), 10) + machineID
}

// String 返回value的json字符串
func (v *Value) String() string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}

// FromString 解析字符串到value
func (v *Value) FromString(s string) error {
	return json.Unmarshal([]byte(s), v)
}

// MakeKey 创建服务key
func MakeKey(firefix, machineID string) string {
	var v Value
	{
		v.ID = ID
		v.Name = Name
		v.Frefix = firefix
		v.MachineID = machineID
	}
	return v.Key()
}
