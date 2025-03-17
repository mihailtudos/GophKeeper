package messages

type LoginSuccessMsg struct {
	Token string
}

type ActionMsg struct {
	Value string
}

type LogoutMsg struct{}
