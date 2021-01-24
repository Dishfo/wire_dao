package user_service

import (
	"github.com/Dishfo/wire_dao/db/user_model"
)

type UserService struct {
}

func (s *UserService) GetUser() {
	userInst := &user_model.UserModel{
		Name:     "xxxxxx",
		Password: "159357ghj",
	}

	userInst.OutPut()
}
