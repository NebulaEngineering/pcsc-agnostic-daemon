package dto

import "encoding/hex"

// {
//     "atr": {
//         "protocolInterfaceA": "00000000",
//         "protocolInterfaceB": "00000000",
//         "protocolInterfaceC": "00000000",
//         "protocolInterfaceD": "80010000",
//         "historicalBytes": "C10521300077C1",
//         "tckValid": true,
//         "raw": "3B878001C10521300077C165"
//     },
//     "statusCode": 2,
//     "status": "Shared"
// }

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

type AtrInfo struct {
}

type Atr struct {
	ProtocolInterfaceA string `json:"protocolInterfaceA"`
	ProtocolInterfaceB string `json:"protocolInterfaceB"`
	ProtocolInterfaceC string `json:"protocolInterfaceC"`
	ProtocolInterfaceD string `json:"protocolInterfaceD"`
	HistoricalBytes    string `json:"historicalBytes"`
	TckValid           bool   `json:"tckValid"`
	Raw                string `json:"raw"`
}

type SmartcardStatus struct {
	ATR        *Atr       `json:"atr"`
	StatusCode StatusCode `json:"statusCode"`
	Status     string     `json:"status"`
}

func NewSmartCardStatus(atr []byte, status StatusCode) *SmartcardStatus {

	s := &SmartcardStatus{}
	if len(atr) >= 5 {

		historicalBytes := make([]byte, 0)
		lenHistoricalBytes := int(atr[1] & 0xF)
		historicalBytes = append(historicalBytes, atr[4:4+lenHistoricalBytes]...)

		a := &Atr{}
		a.Raw = hex.EncodeToString(atr)
		a.HistoricalBytes = hex.EncodeToString(historicalBytes)
		a.TckValid = true
		a.ProtocolInterfaceA = "00000000"
		a.ProtocolInterfaceB = "00000000"
		a.ProtocolInterfaceC = "00000000"
		a.ProtocolInterfaceD = "80010000"

		s.ATR = a
	}

	s.StatusCode = status
	s.Status = status.String()

	return s
}
