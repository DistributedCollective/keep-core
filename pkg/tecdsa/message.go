package tecdsa

import (
	"github.com/keep-network/paillier"
)

// InitMessage is a message payload that carries the sender's share of
// `dsaKeyShare` during T-ECDSA distributed DSA key generation as well as
// proofs of correctness for the shares. Sender's share is encrypted with (t, n)
// Paillier threshold key. The message is expected to be broadcast publicly.
type InitMessage struct {
	secretKeyShare *paillier.Cypher
	publicKeyShare *CurvePoint

	// TODO: add all the required ZKPs
}