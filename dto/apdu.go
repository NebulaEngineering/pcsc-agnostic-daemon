package dto

import (
	"encoding/hex"

	"gitlab.com/nebulaeng/fleet/pcscrest/utils"
)

type APDUResponse struct {
	Cmd      string `json:"apdu"`
	Response string `json:"response"`
	Valid    bool   `json:"isValid"`
}

func NewAPDUResponse(apdu, response []byte) *APDUResponse {
	a := &APDUResponse{}
	a.Cmd = hex.EncodeToString(apdu)
	a.Response = hex.EncodeToString(response)

	if len(apdu) <= 0 || len(response) < 2 {
		return a
	}

	if len(apdu) > 0 && (apdu[0]&0x03 == 0x80 || apdu[0] == 0xFF) {
		if utils.VerifyResponseISO7816(response) {
			a.Valid = true
		}
	} else if utils.VerifyResponse(response) {
		a.Valid = true
	}

	return a
}
