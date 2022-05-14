package utils

func VerifyResponse(response []byte) bool {
	if len(response) < 1 {
		return false
	}
	return response[0] == 0x90 || response[0] == 0xAF || response[0] == 0x00
}

func VerifyResponseISO7816(response []byte) bool {
	if len(response) < 2 {
		return false
	}
	return response[len(response)-2] == 0x90 &&
		(response[len(response)-1] == 0x00 || response[len(response)-1] == 0xAF)
}
