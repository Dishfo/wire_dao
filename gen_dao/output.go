package gen_dao

import "fmt"

func (s *Scanner) ListAllType() {
	for pack, types := range s.allTypes {
		fmt.Printf("under package :%s\n", pack)
		for _, description := range types {
			fmt.Printf("  %s\n",description.TypeName)
		}
		fmt.Println()
	}

}

func (s *Scanner)
