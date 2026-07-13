package main

import (
	"testing"

	"github.com/surge-go/fox"
)

func TestRegisterSwaggerRoutes(t *testing.T) {
	printRoutes := false
	engine := fox.New(&fox.Config{
		Addr:        ":0",
		Mode:        fox.ModeTest,
		PrintRoutes: &printRoutes,
	})

	registerSwaggerRoutes(engine)
}
