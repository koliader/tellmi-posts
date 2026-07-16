package token

import "time"

type Maker interface {
	CreateToken(id int64, role string, duration time.Duration) (string, error)
	VerifyToken(tokenString string) (*Payload, error)
}
