package domain

import (
	"encoding/base64"
	"time"
)

type LoginSecret struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CardDetails struct {
	CardNumber     string `json:"card_number"`
	CardHolder     string `json:"card_holder"`
	ExpirationDate string `json:"expiration_date"`
	CVV            string `json:"cvv"`
}

type PlainText string

type BinaryData struct {
	Data []byte // Field must be exported for proper encoding
}

func (b BinaryData) StringBase64() string {
	return base64.StdEncoding.EncodeToString(b.Data)
}

type Secret struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	SType     string    `json:"s_type"`
	SName     string    `json:"s_name"`
	Data      []byte    `json:"data"`
	IV        []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
