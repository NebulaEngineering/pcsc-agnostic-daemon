package reader

import (
	"github.com/ebfe/scard"
	"gitlab.com/nebulaeng/fleet/pcscrest/internal/pcsc/context"
)

type Reader struct {
	name string
	ctx  *context.Context
}

func (r *Reader) Name() string {
	return r.name
}

func (r *Reader) ConnectCard() (*scard.Card, error) {

	card, err := r.ctx.Connect(r.name, scard.ShareExclusive, scard.ProtocolT1)
	if err != nil {
		return nil, err
	}
	return card, nil
}
