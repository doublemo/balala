// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package proto 协议处理
package proto

import "errors"

// RequestBytes 数据流解析
type RequestBytes struct {
	Ver     int8
	Cmd     Command
	SubCmd  Command
	P       int
	PCount  int
	SeqID   uint32
	Content []byte
}

// V 版本
func (req *RequestBytes) V() int8 {
	return req.Ver
}

// Command 主命令号
func (req *RequestBytes) Command() Command {
	return req.Cmd
}

// SubCommand 子命令号
func (req *RequestBytes) SubCommand() Command {
	return req.SubCmd
}

// SID 请求编号
func (req *RequestBytes) SID() uint32 {
	return req.SeqID
}

// Body 请求编号
func (req *RequestBytes) Body() []byte {
	return req.Content
}

// Marshal 封包
func (req *RequestBytes) Marshal() ([]byte, error) {
	if !req.IsValid() {
		return nil, errors.New("Unexpected data")
	}

	var b BytesBuffer
	b.WriteUint32(req.SeqID)
	b.WriteInt8(int8(req.PCount))
	if req.PCount > 1 {
		if req.P <= 1 {
			b.WriteInt8(req.Ver)
			b.WriteInt16(int16(req.Cmd))
			b.WriteInt16(int16(req.SubCmd))
		}
		b.WriteInt8(int8(req.P))
	} else {
		b.WriteInt8(req.Ver)
		b.WriteInt16(int16(req.Cmd))
		b.WriteInt16(int16(req.SubCmd))
	}

	if err := b.WriteBytes(req.Content...); err != nil {
		return nil, err
	}

	return b.Data(), nil
}

// IsValid 检查数据是否合法
func (req *RequestBytes) IsValid() bool {
	if req.SeqID < 1 || req.Cmd < 1 || req.SubCmd < 1 {
		return false
	}

	if len(req.Content) < 1 {
		return false
	}

	return true
}

// Unmarshal 解析rquest 数据
func (req *RequestBytes) Unmarshal(frame []byte) error {
	rd := NewBytesBuffer(frame)
	sid, err := rd.ReadUint32()
	if err != nil {
		return err
	}

	count, err := rd.ReadInt8()
	if err != nil {
		return err
	}

	req.SeqID = sid
	req.PCount = int(count)
	return nil
}
