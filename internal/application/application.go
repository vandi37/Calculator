package application

import (
	"log"
)

type Application struct {
}

func New() *Application {
	return &Application{}
}

func (a *Application) Run() {
	// The program
	log.Println("the program is working")
	// The program end

	// Returning without error
}
