// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package main

import (
	"log"
	"os"
	"strings"

	"github.com/doublemo/balala/cores/services"
)

// installService 安装服务
func installService(serviceName, dname, description, args string) {
	argsVar := strings.Split(args, ",")
	service := services.ServiceWindowsConfig{
		Name:        serviceName,
		DisplayName: dname,
		Description: description,
		Arguments:   argsVar,
	}

	if err := services.InstallWindowsService(&service); err != nil {
		log.Printf("Install Service: %v\n", err)
	} else {
		log.Println("Install Service: success")
	}

	os.Exit(0)
}

// uninstallService 卸载服务
func uninstallService(serviceName string) {
	if err := services.UninstallWindowsService(serviceName); err != nil {
		log.Printf("Uninstall Service: %v\n", err)
	} else {
		log.Println("Uninstall Service: success")
	}
	os.Exit(0)
}
