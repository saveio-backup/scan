package client

import "github.com/gogo/protobuf/proto"

type RecvMsg struct {
	From    string
	Message proto.Message
}
