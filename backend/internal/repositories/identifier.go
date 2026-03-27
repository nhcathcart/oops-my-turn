package repositories

import "go.jetify.com/typeid"

type UserPrefix struct{}

func (UserPrefix) Prefix() string { return "u" }

type UserID struct {
	typeid.TypeID[UserPrefix]
}
