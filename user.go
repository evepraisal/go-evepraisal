package evepraisal

import (
	"encoding/gob"
)

func init() {
	gob.Register(User{})
}

type User struct {
	CharacterName      string
	CharacterOwnerHash string
}
