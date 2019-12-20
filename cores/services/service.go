// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package services

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Options 服务信息
type Options struct {
	// ID 服务唯一识别ID
	ID int32 `json:"id"`

	// Name 服务唯一识别名称
	Name string `json:"name"`

	// MachineID 当前软件机器码
	MachineID string `json:"mid"`

	// IP 当前IP
	IP string `json:"ip"`

	// Port 当前提供内网服务的端口
	Port string `json:"port"`

	// Params 其它参数
	Params map[string]string `json:"p"`
}

// Run 启动服务
func Run(s Server) error {
	handleSignals(s)
	return runService(s)
}

// RegKey 注册服务需要的Key
func RegKey(frefix string, opts *Options) string {
	key := frefix + "/" + strconv.FormatInt(int64(opts.ID), 10)
	if opts.MachineID != "" {
		key += "/" + opts.MachineID
	}

	return key
}

// RegValue 注册服务值
func RegValue(opts *Options) string {
	bytes, _ := json.Marshal(opts)
	return string(bytes)
}

// RegValueFromString 解析服务值
func RegValueFromString(s string) (*Options, error) {
	opts := Options{}
	if err := json.Unmarshal([]byte(s), &opts); err != nil {
		return nil, err
	}

	return &opts, nil
}

// ExecPath 获取前可执行文件完全路径
func ExecPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}

	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, "\\", "/")
	}
	return path, err
}
