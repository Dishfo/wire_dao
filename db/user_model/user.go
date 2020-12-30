package user_model

import (
	"fmt"
	base1 "git.uidev.tools/unifi/wire_dao/db/base"
)


type UserModel struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (m *UserModel) OutPut() {
	fmt.Println(m.Name, m.Password)
}
func (m UserModel) List(tp *base1.TestType) {
	fmt.Println(m.Name, m.Password)
}


func (m UserModel) Delete(bool,bool) {
	fmt.Println(m.Name, m.Password)
}
