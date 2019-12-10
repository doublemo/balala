// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package services

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Run 启动服务
func Run(s Server) error {
	handleSignals(s)
	return runService(s)
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
