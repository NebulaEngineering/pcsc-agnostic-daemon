package dto

import (
	"encoding/hex"
)

type Reader struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	IdHex string `json:"idHex"`
}

// type Reader interface {
// 	GetName() string
// 	GetID() string
// 	GetIdHex() string
// }

func NewReader(id, name string) *Reader {
	r := &Reader{}
	r.ID = id
	r.Name = name

	r.IdHex = hex.EncodeToString([]byte(id))

	return r
}

// func (r *reader) GetName() string {
// 	return r.Name
// }

// func (r *reader) GetID() string {
// 	return r.ID
// }

// func (r *reader) GetIdHex() string {
// 	return r.IdHex
// }
