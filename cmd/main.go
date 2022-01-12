package main

import (
	"analitics-go/pkg/application"
)

type AppConfig struct {
	application.Config
}

func main() {
	cfg := application.GetConfig()
	application.Run(cfg)
}
