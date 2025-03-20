package domain

import (
	"encoding/base64"
	"errors"
	"time"
)

var (
	ErrInvalidSecret = errors.New("invalid secret")
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

type PlainText struct {
	Value string `json:"value"`
}

type BinaryData struct {
	Data []byte `json:"-"` // We exclude it from JSON marshaling
}

func (b BinaryData) ToBase64() string {
	return base64.StdEncoding.EncodeToString(b.Data)
}

// Convert base64 string back to raw bytes
func (b *BinaryData) FromBase64(encoded string) error {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	b.Data = data
	return nil
}

type Secret struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	SType     string    `json:"s_type"`
	SName     string    `json:"s_name"`
	Data      []byte    `json:"data"`
	IV        []byte    `json:"-"`
	SumCheck  []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
