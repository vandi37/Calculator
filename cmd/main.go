package main

import "github.com/vandi37/Calculator/internal/application"

func main() {
	app := application.New("config/config.json")
	app.Run()
}
