package chesspb

import (
	"bufio"
	"encoding/binary"
	"errors"
	//	"github.com/TheDiscordian/speedychess/chess"
	. "github.com/TheDiscordian/speedychess/flags"
	"google.golang.org/protobuf/proto"
)

// ReadMessage reads a packet from r, and stores the data in m (if valid) or returns an error. Server can
// receive up to 256 bytes, while a client has no receive limit.
func ReadMessage(r *bufio.Reader, m *proto.Message) error {
	mType, err := r.ReadByte()
	if err != nil {
		return err
	}

	*m = NewMsg(mType)
	if *m == nil {
		return errors.New("Invalid message type")
	}
	if mType == PingMsg {
		return nil
	}

	var size int
	if SERVER {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		size = int(b)
	} else {
		size64, err := binary.ReadUvarint(r)
		if err != nil {
			return err
		}
		size = int(size64)
	}
	if size == 0 {
		return nil
	}

	recv := make([]byte, size, size)
	for i := 0; i < size; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		recv[i] = b
	}
	err = proto.Unmarshal(recv, *m)
	return err
}

// BuildMessage converts a Message into a byte slice for sending over the network.
func BuildMessage(m proto.Message) ([]byte, error) {
	mType := byte(m.ProtoReflect().Descriptor().Index())
	if !SERVER && mType == PingMsg {
		return []byte{mType}, nil
	}

	mBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}

	var mSize []byte
	if !SERVER {
		mSize = []byte{byte(len(mBytes))}
	} else {
		mSize = make([]byte, 8)
		mLen := binary.PutUvarint(mSize, uint64(len(mBytes)))
		mSize = mSize[:mLen]
	}

	out := make([]byte, len(mBytes)+len(mSize)+1)
	out[0] = mType
	copy(out[1:], mSize)
	copy(out[1+len(mSize):], mBytes)

	return out, nil
}
