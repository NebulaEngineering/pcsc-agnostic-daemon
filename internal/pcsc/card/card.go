package card

import (
	"errors"
	"fmt"

	"github.com/ebfe/scard"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/reader"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/utils"
)

type Card struct {
	card      *scard.Card
	uid       []byte
	sessionId string
}

type StatusCode int

const (
	Ready      StatusCode = 1
	Shared     StatusCode = 2
	NotPresent StatusCode = -1
)

// String return string with current state's card.
func (s StatusCode) String() string {
	switch s {
	case Ready:
		return "Ready"
	case Shared:
		return "Shared"
	case NotPresent:
		return "NotPresent"
	}
	return ""
}

// ConnectCard verify presence's card on reader and return Card instance.
func ConnectCard(r *reader.Reader) (*Card, error) {
	if r == nil {
		return nil, errors.New("reader is not valid")
	}

	internalCard, err := r.ConnectCard()
	if err != nil {
		return nil, err
	}

	resp, err := internalCard.Transmit([]byte{0xFF, 0xCA, 0x00, 0x00, 0x00})
	if err != nil {
		return nil, err
	} else {
		if len(resp) <= 2 {
			return nil, errors.New("error on get UID")
		}
	}
	uid := make([]byte, len(resp)-2)
	copy(uid, resp)
	c := &Card{card: internalCard, uid: uid}

	if !reader.IsEnForceIso14443_4() {
		if s, err := internalCard.Status(); err != nil {
			return nil, err
		} else {
			if len(s.Atr) > 14 && s.Atr[14] == 0x20 {
				c.TransparentSessionStart()
				c.Switch1444_4()
				c.TransparentSessionEnd()
			}

		}
	}

	return c, nil
}

// String print card features
func (c *Card) String() string {
	status, err := c.card.Status()
	if err != nil {
		return ""
	}
	statusCard, _ := c.Status()

	var historicalBytes []byte
	if len(status.Atr) >= 5 {
		historicalBytes = make([]byte, 0)
		lenHistoricalBytes := int(status.Atr[1] & 0xF)
		historicalBytes = append(historicalBytes, status.Atr[4:4+lenHistoricalBytes]...)
	} else {
		return fmt.Sprintf(`{ status: %s }`, statusCard)
	}

	str := fmt.Sprintf(`{
		status: %s,
		atr: [% 02X],
		historical: [% 02X],
	}`, statusCard, status.Atr, historicalBytes)
	return str
}

// GetUID return UID of card
func (c *Card) GetUID() []byte {
	return c.uid
}

// SetSessionID set session id of card
func (c *Card) SetSessionID(sessionId string) {
	c.sessionId = sessionId
}

// GetSessionID return session id of card
func (c *Card) GetSessionID() string {
	return c.sessionId
}

// SendAPDU send APDU 'data' to card through the reader and wait for a response.
func (c *Card) SendAPDU(data []byte) ([]byte, error) {

	if utils.Debug {
		fmt.Printf("APDU: [% 02X]\n", data)
	}
	response, err := c.card.Transmit(data)
	if err != nil {
		return nil, err
	}
	if utils.Debug {
		fmt.Printf("RESPONSE: [% 02X]\n", response)
	}
	return response, nil
}

// Disconnect release card from reader
func (c *Card) Disconnect() error {
	if err := c.card.Disconnect(scard.LeaveCard); err != nil {
		return err
	}
	return nil
}

// Atr return ATR bytes of card.
func (c *Card) Atr() ([]byte, error) {

	status, err := c.card.Status()
	if err != nil {
		return nil, err
	}

	return status.Atr, nil
}

// Status return status's card on reader
func (c *Card) Status() (StatusCode, error) {

	status, err := c.card.Status()
	if err != nil {
		return NotPresent, err
	}

	switch status.State & 0xFF {
	case scard.Powered | scard.Present | scard.Negotiable:
		return Ready, nil
	case scard.Absent | scard.Present:
		return Ready, nil
	case scard.Powered | scard.Present:
		return Ready, nil
	case scard.Powered | scard.Present | scard.Specific:
		return Ready, nil
	case scard.Absent:
		return NotPresent, nil
	case scard.Swallowed:
		return Shared, nil
	}

	fmt.Printf("statusCode: %X, %X\n", status.State, (status.State & 0xFF))

	return NotPresent, nil
}

// Transparent Session (PCSC)
func (c *Card) TransparentSessionStart() ([]byte, error) {
	apdu := []byte{0xFF, 0xC2, 0x00, 0x00, 0x04, 0x81, 0x00, 0x84, 0x00}
	resp, err := c.SendAPDU(apdu)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// TransparentSessionStartOnly start transparent session to send APDU
func (c *Card) TransparentSessionStartOnly() ([]byte, error) {
	apdu := []byte{0xFF, 0xC2, 0x00, 0x00, 0x02, 0x81, 0x00}
	resp, err := c.SendAPDU(apdu)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// TransparentSessionResetRF start transparent session to send APDU
func (c *Card) TransparentSessionResetRF() ([]byte, error) {
	apdu1 := []byte{0xFF, 0xC2, 0x00, 0x00, 0x02, 0x83, 0x00}
	resp, err := c.SendAPDU(apdu1)
	if err != nil {
		return resp, err
	}
	apdu2 := []byte{0xFF, 0xC2, 0x00, 0x00, 0x02, 0x84, 0x00}
	resp2, err := c.SendAPDU(apdu2)
	if err != nil {
		return resp2, err
	}
	return resp2, nil
}

// TransparentSessionEnd finish transparent session
func (c *Card) TransparentSessionEnd() ([]byte, error) {
	apdu := []byte{0xFF, 0xC2, 0x00, 0x00, 0x02, 0x82, 0x00, 0x00}
	resp, err := c.SendAPDU(apdu)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// Switch1444_4 switch channel reader to send ISO 1444-4 APDU
func (c *Card) Switch1444_4() ([]byte, error) {
	apdu := []byte{0xff, 0xc2, 0x00, 0x02, 0x04, 0x8F, 0x02, 0x00, 0x04}
	resp, err := c.SendAPDU(apdu)
	if err != nil {
		return resp, err
	}
	return resp, nil
}
