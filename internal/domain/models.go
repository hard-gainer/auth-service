package domain

type User struct {
	ID       int64
	Name     string
	Email    string
	PassHash []byte
	Role     string
}

type UserInfo struct {
	ID    int64
	Name  string
	Email string
	Role  string
}

type App struct {
	ID     int
	Name   string
	Secret string
}
