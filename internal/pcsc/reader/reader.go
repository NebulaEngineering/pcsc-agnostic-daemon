package reader

import (
	"github.com/ebfe/scard"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/context"
)

type Reader struct {
	name string
	ctx  *context.Context
}

//Name return name's reader
func (r *Reader) Name() string {
	return r.name
}

//ConnectCard connect card in reader. Protocol used is T=1.
func (r *Reader) ConnectCard() (*scard.Card, error) {

	card, err := r.ctx.Connect(r.name, scard.ShareExclusive, scard.ProtocolT1)
	if err != nil {
		return nil, err
	}
	return card, nil
}
