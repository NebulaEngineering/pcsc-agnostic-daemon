package reader

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/ebfe/scard"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/context"
)

var forceIso14443_3 bool

func init() {
	flag.BoolVar(&forceIso14443_3, "force-iso14443-3", false, "Force ISO 14443-4 mode")
}

// ConnectReader verify reader, configure reader to send automatic polling for card detyection and return reader instance.
func ConnectReader(c *context.Context, reader string) (*Reader, error) {
	if c == nil {
		if ctx, err := context.New(); err != nil {
			return nil, err
		} else {
			c = ctx
		}
	}

	if ok, err := c.IsValid(); err != nil || !ok {
		return nil, fmt.Errorf("context is not valid, err: %w", err)
	}
	readers, err := c.ListReaders()
	if err != nil {
		return nil, err
	}
	flag := false
	newReader := ""
	for _, r := range readers {
		if strings.Contains(r, reader) {
			newReader = r
			flag = true
			break
		}
	}
	if !flag {
		return nil, errors.New("reader not found")
	}

	if strings.Contains(newReader, "PICC") {
		direct, err := c.Connect(newReader, scard.ShareDirect, scard.ProtocolT0)
		if err != nil {
			return nil, err
		}
		resp1, err := direct.Control(0x42000000+2079, []byte{0x23, 0x00})
		if err != nil {
			return nil, err
		}
		if len(resp1) > 0 && resp1[len(resp1)-1] != (0x8F&func() byte {
			if forceIso14443_3 {
				return 0x7F
			}
			return 0xFF
		}()) {
			if _, err := direct.Control(0x42000000+2079, []byte{0x23, 0x01, 0x8F}); err != nil {
				return nil, err
			}
		}
		direct.Disconnect(scard.LeaveCard)
	}

	return &Reader{
		name: newReader,
		ctx:  c,
	}, nil

}
