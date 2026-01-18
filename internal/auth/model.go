package auth

import "encoding/json"

type Token string

func (m Token) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

type UserInfo struct {
	Id       int32
	Username string
	Role     string
}

type TokenPair struct {
	Access  Token
	Refresh Token
}
