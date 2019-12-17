// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package service 服务
package service

// MakeKey 服务KEY,用于ETCD服务发现
func MakeKey(serviceName, frefix, idaddress string) string {
	if idaddress != "" {
		idaddress = "/" + idaddress
	}

	return frefix + "/" + serviceName + idaddress
}
