package evepraisal

import (
	"encoding/gob"
)

func init() {
	gob.Register(User{})
}

// User is information about a logged-in user. Currently, this is only stored in the user's session but may be used
// as keys in a user settings database
type User struct {
	CharacterName      string
	CharacterOwnerHash string
}
