package domain

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Salt     string `json:"-"`
	Password string `json:"password"`
}