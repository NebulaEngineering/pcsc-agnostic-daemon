package card

import (
	"errors"
	"fmt"

	"github.com/ebfe/scard"
	"github.com/nebulaengineering/pcsc-agnostic-daemon/internal/pcsc/reader"
)

type Card struct {
	card *scard.Card
}

type StatusCode int

const (
	Ready      StatusCode = 1
	Shared     StatusCode = 2
	NotPresent StatusCode = -1
)

//String return string with current state's card.
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

//ConnectCard verify presence's card on reader and return Card instance.
func ConnectCard(r *reader.Reader) (*Card, error) {
	if r == nil {
		return nil, errors.New("reader is not valid")
	}

	internalCard, err := r.ConnectCard()
	if err != nil {
		return nil, err
	}
	return &Card{card: internalCard}, nil
}

//String print card features
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
		atr: [% X],
		historical: [% X],
	}`, statusCard, status.Atr, historicalBytes)
	return str
}

//SendAPDU send APDU 'data' to card through the reader and wait for a response.
func (c *Card) SendAPDU(data []byte) ([]byte, error) {

	response, err := c.card.Transmit(data)
	if err != nil {
		return nil, err
	}
	return response, nil
}

//Disconnect release card from reader
func (c *Card) Disconnect() error {
	if err := c.card.Disconnect(scard.ResetCard); err != nil {
		return err
	}
	return nil
}

//Atr return ATR bytes of card.
func (c *Card) Atr() ([]byte, error) {

	status, err := c.card.Status()
	if err != nil {
		return nil, err
	}

	return status.Atr, nil
}

//Status return status's card on reader
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
