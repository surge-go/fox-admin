package main

import (
	"flag"
	"log"

	"fox-admin/internal/application"
	"fox-admin/internal/module/system"

	"github.com/surge-go/fox/middleware"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	app, err := application.New(*configPath)
	if err != nil {
		log.Fatalf("init application: %v", err)
	}

	defer app.Close()
	app.Engine().Use(
		middleware.Tracing(),
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Timeout(),
		middleware.CORS(),
		middleware.BodyLimit(),
		middleware.RateLimit(),
		middleware.Gzip(),
	)
	v1 := app.Engine().Group("/api/v1")
	system.RegisterRoutes(v1, app.DB(), app.Logger())
	if err = app.Run(); err != nil {
		log.Fatalf("run application: %v", err)
	}
}
