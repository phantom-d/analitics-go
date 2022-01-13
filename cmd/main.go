package main

import (
	"analitics-go/pkg/application"
)

type AppConfig struct {
	application.Application
}

func main() {
	app := application.New()
	app.Run()
}
