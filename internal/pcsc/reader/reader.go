package reader

import (
	"flag"

	"fmt"

	"github.com/ebfe/scard"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/context"
)

var forceIso14443_4 bool
var forceIso14443_3 bool

func init() {
	flag.BoolVar(&forceIso14443_3, fmt.Sprint("\x1B[9m", "force-iso14443-3", "\x1B[0m"), false,
		fmt.Sprint("\x1B[9m", "Force ISO 14443-3 mode", "\x1B[0m", " DEPRECATED"))
}

type Reader struct {
	name string
	ctx  *context.Context
}

func IsEnForceIso14443_4() bool {
	return forceIso14443_4
}

// Name return name's reader
func (r *Reader) Name() string {
	return r.name
}

// ConnectCard connect card in reader. Protocol used is T=1.
func (r *Reader) ConnectCard() (*scard.Card, error) {

	card, err := r.ctx.Connect(r.name, scard.ShareExclusive, scard.ProtocolT1)
	if err != nil {
		return nil, err
	}
	return card, nil
}
