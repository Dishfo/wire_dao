package main

import (
	"context"
	"fmt"
	"golang.org/x/tools/go/packages"
	"os"
)

func main() {
	ctx := context.TODO()
	wd, _ := os.Getwd()
	cfg := &packages.Config{
		Context:    ctx,
		Mode:       packages.LoadAllSyntax,
		Dir:        wd+"/service/user_service",
		Env:        os.Environ(),
		BuildFlags: []string{"-tags=wireinject"},
	}

	loadPackages,err := packages.Load(cfg)

	if err != nil {
		fmt.Println("load failed ",err.Error())
	}

	for _, loadPackage := range loadPackages {
		fmt.Println(loadPackage.Name)
	}



}
