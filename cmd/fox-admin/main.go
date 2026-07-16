package main

import (
	"flag"
	"io"
	"log"
	"net/http"

	"fox-admin/internal/application"
	"fox-admin/internal/middleware"
	"fox-admin/internal/module/system"

	"github.com/surge-go/fox"
	foxmiddleware "github.com/surge-go/fox/middleware"
	"github.com/wdcbot/qingfeng"
)

// @title fox-admin API
// @version 1.0.0
// @description fox-admin 后台管理接口文档
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	app, err := application.New(*configPath)
	if err != nil {
		log.Fatalf("init application: %v", err)
	}

	defer app.Close()
	corsConfig := newCORSConfig()
	app.Engine().Use(
		foxmiddleware.Tracing(),
		foxmiddleware.RequestID(),
		foxmiddleware.Logger(),
		foxmiddleware.Timeout(),
		foxmiddleware.CORSWithConfig(corsConfig),
		foxmiddleware.BodyLimit(),
		foxmiddleware.RateLimit(),
		foxmiddleware.Gzip(),
	)
	authManager := app.AuthManager()
	if authManager == nil {
		log.Fatal("auth manager is nil")
	}
	v1 := app.Engine().Group("/api/v1", middleware.AuthWithConfig(middleware.AuthConfig{
		Manager: authManager,
		SkipPaths: []string{
			"/api/v1/system/auth/login",
			"/api/v1/system/auth/refresh",
		},
	}))
	closeSystem := system.RegisterRoutes(v1, app.DB(), authManager, app.Logger())
	defer closeSystem()
	registerSwaggerRoutes(app.Engine())
	if err = app.Run(); err != nil {
		log.Fatalf("run application: %v", err)
	}
}

func newCORSConfig() foxmiddleware.CORSConfig {
	corsConfig := foxmiddleware.DefaultCORSConfig()
	corsConfig.AllowHeaders = append(
		corsConfig.AllowHeaders,
		middleware.DefaultDeviceIDHeaderName,
		middleware.DefaultRefreshHeaderName,
	)
	corsConfig.ExposeHeaders = []string{
		"X-Request-ID",
		middleware.DefaultAccessResponseHeaderName,
		middleware.DefaultRefreshResponseHeaderName,
		middleware.AccessExpiresAtHeaderName,
		middleware.RefreshExpiresAtHeaderName,
		middleware.TokenTypeHeaderName,
	}
	return corsConfig
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
