package main

import (
	"flag"
	"io"
	"log"
	"net/http"

	"fox-admin/internal/application"
	"fox-admin/internal/module/system"

	"github.com/surge-go/fox"
	"github.com/surge-go/fox/middleware"
	"github.com/wdcbot/qingfeng"
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
	system.RegisterRoutes(v1, app.DB(), app.AuthManager(), app.Logger())
	registerSwaggerRoutes(app.Engine())
	if err = app.Run(); err != nil {
		log.Fatalf("run application: %v", err)
	}
}

func registerSwaggerRoutes(engine *fox.Engine) {
	swaggerHandler := newSwaggerHandler(qingfeng.Config{
		Title:       "fox-admin API",
		Description: "fox-admin 后台管理接口文档",
		Version:     "1.0.0",
		BasePath:    "/docs",
		DocPath:     "./docs/swagger.json",
		EnableDebug: true,
		UITheme:     qingfeng.ThemeModern,
	})

	// 青峰 Swag 提供标准 http.Handler，这里通过 fox.Context 暴露的原始请求和响应完成适配。
	handleSwagger := func(c *fox.Context) {
		swaggerHandler.ServeHTTP(c.RawWriter(), c.RawRequest())
	}

	engine.GET("/docs", handleSwagger)
	engine.GET("/docs/*filepath", handleSwagger)
}

func newSwaggerHandler(cfg qingfeng.Config) http.Handler {
	originalOutput := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalOutput)

	// 青峰 Swag 当前版本会在构造 HTTPHandler 时无条件打印 banner，这里只静默构造阶段日志。
	return qingfeng.HTTPHandler(cfg)
}
