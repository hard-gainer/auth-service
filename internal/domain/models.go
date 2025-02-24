package domain

type User struct {
	ID       int64
	Name     string
	Email    string
	PassHash []byte
}

type UserInfo struct {
	ID       int64
	Name     string
	Email    string
}

type App struct {
	ID     int
	Name   string
	Secret string
}
