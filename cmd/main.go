package main

import (
	"log/slog"

	"go.uber.org/zap"
)

type Repo struct {
	users []*User
}
type User struct {
	Email    string
	Password string
	Token    string
}

func (r *Repo) login(u User, z *zap.Logger) {
	slog.Info("Login started")

	r.users = append(r.users, &u)

	slog.Info("auth successes with user",
		"email", u.Email,
		"password", u.Password,
		"token", u.Token,
	)

	z.Info("login ended 🥳")
}

func main() {
	z := zap.NewNop()
	r := &Repo{}
	u := User{Email: "a@b.c", Password: "p", Token: "t"}
	r.login(u, z)
}
