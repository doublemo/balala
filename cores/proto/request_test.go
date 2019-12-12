package proto

import (
	"reflect"
	"testing"
)

func TestRequestBytes(t *testing.T) {
	rq := &RequestBytes{}
	rq.Ver = 1
	rq.Cmd = 10001
	rq.SubCmd = 12
	rq.P = 1
	rq.PCount = 1
	rq.SeqID = 154
	rq.Content = make([]byte, 10)

	bytes0, err := rq.Marshal()
	if err != nil {
		t.Fatal(err)
		return
	}

	// 带分页
	rq.PCount = 10
	bytes1, err := rq.Marshal()
	if err != nil {
		t.Fatal(err)
		return
	}

	// 分页内页
	rq.P = 2
	bytes2, err := rq.Marshal()
	if err != nil {
		t.Fatal(err)
		return
	}

	rd2 := &RequestBytes{}
	if err := rd2.Unmarshal(bytes0); err != nil {
		t.Fatal(err)
		return
	}

	rd3 := &RequestBytes{}
	if err := rd3.Unmarshal(bytes1); err != nil {
		t.Fatal(err)
		return
	}

	rd4 := &RequestBytes{}
	if err := rd4.Unmarshal(bytes2); err != nil {
		t.Fatal(err)
		return
	}

	b0, _ := rd2.Marshal()
	b1, _ := rd3.Marshal()
	b2, _ := rd4.Marshal()
	if !reflect.DeepEqual(bytes0, b0) {
		t.Fatalf("Not Equal:\nReceived: '%+v'\nExpected: '%+v'\n", bytes0, b0)
		return
	}

	if !reflect.DeepEqual(bytes1, b1) {
		t.Fatalf("Not Equal:\nReceived: '%+v'\nExpected: '%+v'\n", bytes1, b1)
		return
	}

	if !reflect.DeepEqual(bytes2, b2) {
		t.Fatalf("Not Equal:\nReceived: '%+v'\nExpected: '%+v'\n", bytes2, b2)
		return
	}
}
