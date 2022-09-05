package sampool

import (
	"errors"
	"github.com/0xPolygon/polygon-edge/rootchain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSAMPool_AddMessage(t *testing.T) {
	t.Parallel()

	t.Run(
		"bad hash",
		func(t *testing.T) {
			t.Parallel()

			verifier := mockVerifier{
				verifyHash: func(msg rootchain.SAM) error {
					return errors.New("asdasd")
				},
			}

			pool := New(verifier)

			msg := rootchain.SAM{
				Hash: []byte("some really bad hash"),
			}

			err := pool.AddMessage(msg)

			assert.Error(t, err)
		},
	)

	t.Run(
		"bad signature",
		func(t *testing.T) {
			t.Parallel()

			verifier := mockVerifier{
				verifyHash: func(sam rootchain.SAM) error {
					return nil
				},

				verifySignature: func(sam rootchain.SAM) error {
					return errors.New("some really bad signature")
				},
			}

			pool := New(verifier)

			msg := rootchain.SAM{
				Signature: []byte("some really bad signature"),
			}

			err := pool.AddMessage(msg)

			assert.Error(t, err)
		},
	)

	t.Run(
		"reject stale message",
		func(t *testing.T) {
			t.Parallel()

			verfier := mockVerifier{
				verifyHash: func(sam rootchain.SAM) error {
					return nil
				},
				verifySignature: func(sam rootchain.SAM) error {
					return nil
				},
				quorumFunc: nil,
			}

			pool := New(verfier)
			pool.lastProcessedMessage = 10

			msg := rootchain.SAM{
				Event: rootchain.Event{
					Number: 5,
				},
			}

			err := pool.AddMessage(msg)

			assert.ErrorIs(t, err, ErrStaleMessage)
		},
	)
}