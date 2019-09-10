package common

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/saveio/max/thirdparty/assert"
)

func TestTLV(t *testing.T) {
	descr1 := []byte("This is a test description.")
	tlv1 := NewTLV(TLV_TAG_SIGNATURE, descr1)
	tlv1.Print()

	descr2 := []byte("This is another test description.")
	tlv2 := NewTLV(TLV_TAG_PUBLIC_KEY, descr2)
	tlv2.Print()

	var buf1, buf2 bytes.Buffer
	err := tlv1.Write(&buf1)
	fmt.Println("Buffer1:", buf1.Bytes(), err)

	err = tlv2.Write(&buf2)
	fmt.Println("Buffer2:", buf2.Bytes(), err)

	var options []byte
	options = append(options, buf1.Bytes()...)
	options = append(options, buf2.Bytes()...)
	fmt.Println(options)

	buf_ := bytes.NewBuffer(options)
	tlv_, err := Read(buf_)
	assert.Nil(err, t)
	tlv_.Print()

	tlv_, err = Read(buf_)
	assert.Nil(err, t)
	tlv_.Print()
}
