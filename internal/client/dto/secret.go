package dto

type SecretMessage struct {
	Data           any    `json:"data"`
	MasterPassword string `json:"master_password"`
	Type           string `json:"type"`
	Name           string `json:"name"`
}

type LoginSecret struct {
	Username string `json:"login"`
	Password string `json:"password"`
}
