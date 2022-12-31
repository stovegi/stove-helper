package main

import (
	"context"

	"github.com/stovegi/stove-helper/pkg/config"
	"github.com/stovegi/stove-helper/pkg/helper"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()
	svc, err := helper.NewService(ctx, cfg)
	if err != nil {
		panic(err)
	}
	if err := svc.Start(); err != nil {
		panic(err)
	}
}
