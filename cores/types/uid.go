// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

package types

import (
	"encoding/base64"
	"strconv"

	"github.com/doublemo/balala/cores/crypto/id"
)

// UID 识别ID
type UID uint64

// Uint64 uint64类型转换
func (uid UID) Uint64() uint64 {
	return uint64(uid)
}

// String uid 字符串
func (uid UID) String() string {
	str := strconv.FormatUint(uid.Uint64(), 10)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(str))
}

// Encrypt 加密
func (uid UID) Encrypt(key []byte) (string, error) {
	return id.Encrypt(uid.Uint64(), key)
}

// Decrypt 解密
func (uid *UID) Decrypt(str string, key []byte) error {
	id, err := id.Decrypt(str, key)
	if err != nil {
		return err
	}

	*uid = UID(id)
	return nil
}
