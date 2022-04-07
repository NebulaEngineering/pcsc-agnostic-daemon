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

func NewAPDUResponse(apdu, response []byte, is7816 bool) *APDUResponse {
	a := &APDUResponse{}
	a.Cmd = hex.EncodeToString(apdu)
	a.Response = hex.EncodeToString(response)

	if len(apdu) <= 0 || len(response) < 2 {
		return a
	}

	if is7816 {
		a.Valid = utils.VerifyResponseISO7816(response)
	} else {
		a.Valid = utils.VerifyResponse(response)
	}

	return a
}
