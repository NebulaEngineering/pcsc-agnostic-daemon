package reader

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ebfe/scard"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/context"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/utils"
)

// ConnectReader verify reader and return reader instance.
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

	return &Reader{
		name: newReader,
		ctx:  c,
	}, nil

}

// ConnectReader verify reader, configure reader to send automatic polling for card detyection and return reader instance.
func PrepareReader(c *context.Context, reader string) error {
	if c == nil {
		return errors.New("context is nil")
	}

	if ok, err := c.IsValid(); err != nil || !ok {
		return fmt.Errorf("context is not valid, err: %w", err)
	}
	readers, err := c.ListReaders()
	if err != nil {
		return err
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
		return errors.New("reader not found")
	}

	// if strings.Contains(newReader, "PICC") {
	direct, err := c.Connect(newReader, scard.ShareDirect, scard.ProtocolUndefined)
	if err != nil {
		return err
	}
	apdu := []byte{0x23, 0x00}
	if utils.Debug {
		fmt.Printf("CONTROL APDU: [% 02X]\n", apdu)
	}

	ctlCode := scard.CtlCode(2079)
	//resp1, err := direct.Control(3500, apdu)
	resp1, err := direct.Control(ctlCode, apdu)
	if err != nil {
		return err
	}
	if utils.Debug {
		fmt.Printf("RESPONSE CONTROL APDU: [% 02X]\n", resp1)
	}

	if len(resp1) > 0 && resp1[len(resp1)-1] != (0x8F&func() byte {
		if !forceIso14443_4 {
			return 0x7F
		}
		return 0xFF
	}()) {
		apdu := []byte{0x23, 0x01, 0x8F & func() byte {
			if !forceIso14443_4 {
				return 0x7F
			}
			return 0xFF
		}()}
		if utils.Debug {
			fmt.Printf("CONTROL APDU: [% 02X]\n", apdu)
		}
		if resp, err := direct.Control(ctlCode, apdu); err != nil {
			return err
		} else if utils.Debug {
			fmt.Printf("RESPONSE CONTROL APDU: [% 02X]\n", resp)
		}
	}
	direct.Disconnect(scard.LeaveCard)
	// }

	return nil

}
