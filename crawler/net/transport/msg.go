package transport

import (
	"fmt"
	"io"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
)

type Message interface {
	Code() int
}

type Msg struct {
	code       uint64
	Size       uint32 // Size of the raw payload
	Payload    io.Reader
	ReceivedAt int64
}

// See the eth rlp decode rules
func (msg Msg) Decode(val interface{}) error {
	s := rlp.NewStream(msg.Payload, uint64(msg.Size))
	if err := s.Decode(val); err != nil {
		return fmt.Errorf("invalid msg (code %x) (size %d) %v", msg.Code(), msg.Size, err)
	}
	return nil
}

func (msg Msg) String() string {
	return fmt.Sprintf("msg #%v (%v bytes)", msg.Code, msg.Size)
}

// Discard reads any remaining payload data into a black hole.
func (msg Msg) Discard() error {
	_, err := io.Copy(io.Discard, msg.Payload)
	return err
}

func (msg Msg) Time() time.Time {
	return time.UnixMilli(msg.ReceivedAt)
}

func (msg Msg) Code() int { return int(msg.code) }
