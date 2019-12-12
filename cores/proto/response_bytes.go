// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/balala>

// Package proto 协议处理
package proto

import (
	"errors"
	"strconv"

	"github.com/golang/protobuf/proto"
)

// ResponseBytes 数据流解析
type ResponseBytes struct {
	Ver     int8
	Cmd     Command
	SubCmd  Command
	P       int
	PCount  int
	SeqID   uint32
	Content []byte
	Err     error
}

// V 版本
func (resp *ResponseBytes) V() int8 {
	return resp.Ver
}

// Command 主命令号
func (resp *ResponseBytes) Command() Command {
	return resp.Cmd
}

// SubCommand 子命令号
func (resp *ResponseBytes) SubCommand() Command {
	return resp.SubCmd
}

// SID 请求编号
func (resp *ResponseBytes) SID() uint32 {
	return resp.SeqID
}

// Body 请求编号
func (resp *ResponseBytes) Body() []byte {
	return resp.Content
}

// Marshal 封包
func (resp *ResponseBytes) Marshal() ([]byte, error) {
	if resp.IsError() {
		if resp.SubCmd != InternalBad {
			bad := &Bad{
				Command:    int32(resp.Cmd),
				SubCommand: int32(resp.SubCmd),
				Code:       0,
				Message:    "",
			}

			if m, err := strconv.ParseInt(resp.Err.Error(), 10, 64); err == nil {
				bad.Code = int32(m)
			}

			bad.Message = resp.Err.Error()
			resp.SubCmd = InternalBad
			resp.Content, _ = proto.Marshal(bad)
		}
	}

	if !resp.IsValid() {
		return nil, errors.New("Unexpected data")
	}

	var b BytesBuffer
	b.WriteUint32(resp.SeqID)
	b.WriteInt8(int8(resp.PCount))
	if resp.PCount > 1 {
		b.WriteInt8(int8(resp.P))
		if resp.P <= 1 {
			b.WriteInt8(resp.Ver)
			b.WriteInt16(int16(resp.Cmd))
			b.WriteInt16(int16(resp.SubCmd))
		}
	} else {
		b.WriteInt8(resp.Ver)
		b.WriteInt16(int16(resp.Cmd))
		b.WriteInt16(int16(resp.SubCmd))
	}

	if err := b.WriteBytes(resp.Content...); err != nil {
		return nil, err
	}
	return b.Data(), nil
}

// IsValid 检查数据是否合法
func (resp *ResponseBytes) IsValid() bool {
	if resp.SeqID < 1 {
		return false
	}

	if resp.PCount <= 1 {
		if resp.Cmd < 1 || resp.SubCmd < 1 {
			return false
		}
	}

	if len(resp.Content) < 1 {
		return false
	}

	return true
}

// Unmarshal 解析rquest 数据
func (resp *ResponseBytes) Unmarshal(frame []byte) error {
	rd := NewBytesBuffer(frame)
	sid, err := rd.ReadUint32()
	if err != nil {
		return err
	}

	count, err := rd.ReadInt8()
	if err != nil {
		return err
	}

	if count > 1 {
		page, err := rd.ReadInt8()
		if err != nil {
			return err
		}
		resp.P = int(page)
	}

	if resp.P <= 1 {
		v, err := rd.ReadInt8()
		if err != nil {
			return err
		}

		cmd, err := rd.ReadInt16()
		if err != nil {
			return err
		}

		subcmd, err := rd.ReadInt16()
		if err != nil {
			return err
		}

		resp.Ver = v
		resp.Cmd = Command(cmd)
		resp.SubCmd = Command(subcmd)
	}

	resp.SeqID = sid
	resp.PCount = int(count)
	resp.Content = rd.Bytes()
	return nil
}

// IsError 确认是否有错误
func (resp *ResponseBytes) IsError() bool {
	return resp.Err != nil || resp.SubCmd == InternalBad
}

// Error 获取错误信息
func (resp *ResponseBytes) Error() error {
	return resp.Err
}
