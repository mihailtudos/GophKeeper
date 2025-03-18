package dto

type SecretMessage struct {
	Value any    `json:"value"`
	SType string `json:"s_type"`
	SName string `json:"s_name"`
}

type LoginSecret struct {
	Username string `json:"login"`
	Password string `json:"password"`
}
