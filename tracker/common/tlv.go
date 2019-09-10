package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type TLV_TAG int32

const (
	TLV_TAG_PUBLIC_KEY TLV_TAG = iota
	TLV_TAG_SIGNATURE
)

var (
	ErrTLVRead     = fmt.Errorf("TLV read error")
	ErrTLVWrite    = fmt.Errorf("TLV write error")
	ErrTagNotFound = fmt.Errorf("tag not found")
)

const (
	DEFAULT_TLV_LENGTH  = 4
	MAX_TLV_VALUE_LEGTH = 10000
)

type TLV interface {
	GetTag() TLV_TAG
	GetLength() int
	GetValue() []byte
	Write(w io.Writer) error
	Print()
}

type Record struct {
	Tag    TLV_TAG
	Length int
	Value  []byte
}

func (t *Record) GetTag() TLV_TAG {
	return t.Tag
}

func (t *Record) GetLength() int {
	return t.Length
}

func (t *Record) GetValue() []byte {
	return t.Value
}

func (t *Record) Print() {
	fmt.Printf("Tag: %d\n", t.Tag)
	fmt.Printf("Length: %d\n", t.Length)
	fmt.Printf("Value: %v\n", t.Value)
}

func NewTLV(tag TLV_TAG, value []byte) TLV {
	tlv := new(Record)
	tlv.Tag = tag
	tlv.Length = len(value)
	tlv.Value = make([]byte, tlv.Length)
	copy(tlv.Value, value)
	return tlv
}

func (t *Record) Write(w io.Writer) (err error) {
	err = binary.Write(w, binary.BigEndian, int32(t.GetTag()))
	if err != nil {
		return
	}

	err = binary.Write(w, binary.BigEndian, int32(t.GetLength()))
	if err != nil {
		return
	}

	n, err := w.Write(t.GetValue())
	if err != nil {
		return err
	} else if n != t.Length || t.Length > MAX_TLV_VALUE_LEGTH {
		return ErrTLVWrite
	}
	return
}

func Parse(b []byte) (tlv *Record, err error) {
	buf := bytes.NewBuffer(b)
	return Read(buf)
}

func Read(r io.Reader) (*Record, error) {
	tlv := new(Record)

	var n int32
	err := binary.Read(r, binary.BigEndian, &n)
	if err != nil {
		return tlv, err
	}
	tlv.Tag = TLV_TAG(n)

	err = binary.Read(r, binary.BigEndian, &n)
	if err != nil {
		return tlv, err
	}
	tlv.Length = int(n)

	tlv.Value = make([]byte, tlv.GetLength())
	l, err := r.Read(tlv.Value)
	if err != nil {
		return tlv, err
	} else if l != tlv.GetLength() {
		return tlv, ErrTLVRead
	}
	return tlv, nil
}
