package transport

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
)

type Transport struct {
	*Conn
	Name    string
	Version uint64
	Caps    []p2p.Cap
}

func (t *Transport) Setup(name string, caps []p2p.Cap) error {
	key := crypto.FromECDSAPub(&t.ourKey.PublicKey)[1:]
	if caps == nil {
		return errors.New("0 len caps")
	}
	if err := t.Conn.Write(&Hello{Version: 4, Name: name, Caps: caps, ID: key}); err != nil {
		return nil
	}

	msg, err := t.Conn.Read()
	if err != nil {
		return err
	}

	if msg.code != uint64((Hello{}).Code()) {
		return fmt.Errorf("not helo. got: %v", msg.code)
	}

	hello := new(Hello)
	if err = msg.Decode(hello); err != nil {
		return fmt.Errorf("hello failure. error: %v", err)
	}

	t.Name = hello.Name
	t.Version = hello.Version
	t.Caps = hello.Caps
	return nil
}

func (t *Transport) Read() (Msg, error) {
	message, err := t.Conn.Read()
	if err != nil {
		return Msg{}, err
	}

	switch message.Code() {
	case (Disconnect{}).Code():
		msg := new(Disconnect)
		if err := message.Decode(msg); err != nil {
			return Msg{}, fmt.Errorf("disconnect decode failure: %v", err)
		}
		return Msg{}, msg.Reason
	default:
		return message, nil
	}
}
