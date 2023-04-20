package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type EncPubkey [64]byte

func EncodePubkey(key *ecdsa.PublicKey) EncPubkey {
	var e EncPubkey
	math.ReadBits(key.X, e[:len(e)/2])
	math.ReadBits(key.Y, e[len(e)/2:])
	return e
}

func DecodePubkey(curve elliptic.Curve, e []byte) (*ecdsa.PublicKey, error) {
	if len(e) != len(EncPubkey{}) {
		return nil, errors.New("wrong size public key data")
	}
	p := &ecdsa.PublicKey{Curve: curve, X: new(big.Int), Y: new(big.Int)}
	half := len(e) / 2
	p.X.SetBytes(e[:half])
	p.Y.SetBytes(e[half:])
	if !p.Curve.IsOnCurve(p.X, p.Y) {
		return nil, errors.New("invalid curve point")
	}
	return p, nil
}

func (e EncPubkey) Id() enode.ID {
	return enode.ID(crypto.Keccak256Hash(e[:]))
}
