package v4

import "errors"

// Errors
var (
	ErrUnknownPacket    = errors.New("unknown packet")
	ErrWrongPacket      = errors.New("wrong packet")
	ErrExpired          = errors.New("expired")
	ErrUnsolicitedReply = errors.New("unsolicited reply")
	ErrUnknownNode      = errors.New("unknown node")
	ErrLowPort          = errors.New("low port")
	ErrBondFailure      = errors.New("bond failure")
	ErrHandlerDisabled  = errors.New("handler disabled")
)

func IsImportant(err error) bool {
	if errors.Is(err, ErrHandlerDisabled) {
		return false
	}
	if errors.Is(err, ErrWrongPacket) {
		return false
	}
	if errors.Is(err, ErrUnknownPacket) {
		return false
	}
	if errors.Is(err, ErrExpired) {
		return false
	}

	return true
}
