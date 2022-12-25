package main

import (
	"github.com/stovegi/stove-helper/pkg/config"
	"github.com/stovegi/stove-helper/pkg/helper"
)

func main() {
	service, err := helper.NewService(config.LoadConfig())
	if err != nil {
		panic(err)
	}
	if err := service.Start(); err != nil {
		panic(err)
	}
}
