package card

import (
	"errors"
	"fmt"
	"log"

	"github.com/ebfe/scard"
	"gitlab.com/nebulaeng/fleet/pcscrest/internal/pcsc/reader"
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

func (c *Card) SendAPDU(data []byte) ([]byte, error) {

	response, err := c.card.Transmit(data)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Card) Disconnect() error {

	if err := c.card.Disconnect(scard.LeaveCard); err != nil {
		return err
	}
	return nil
}

func (c *Card) Atr() ([]byte, error) {

	status, err := c.card.Status()
	if err != nil {
		return nil, err
	}

	return status.Atr, nil
}

func (c *Card) Status() (StatusCode, error) {

	status, err := c.card.Status()
	if err != nil {
		return NotPresent, err
	}

	switch status.State & 0xFF {
	case scard.Powered | scard.Present | scard.Negotiable,
		scard.Absent | scard.Present:
		return Ready, nil
	case scard.Absent:
		return NotPresent, nil
	case scard.Swallowed, (scard.Powered | scard.Present):
		return Shared, nil
	}

	log.Printf("statusCode: %X, %X", status.State, (status.State & 0xFF))

	return NotPresent, nil
}
