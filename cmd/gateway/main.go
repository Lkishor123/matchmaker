package main

import (
	"matchmaker/internal/config"
	"matchmaker/internal/handlers"
	"matchmaker/internal/logging"
)

func main() {
	logging.Init()
	cfg, err := config.LoadGateway()
	if err != nil {
		logging.Log.Fatal(err)
	}
	gw, err := handlers.NewGateway(cfg.AuthServiceURL, cfg.UserServiceURL, cfg.MatchServiceURL, cfg.ChatServiceURL, cfg.JWTPrivateKey, 8)
	if err != nil {
		logging.Log.Fatal(err)
	}

	r := logging.NewGinEngine()
	r.GET("/ping", handlers.Ping)
	r.Any("/api/v1/auth/*proxyPath", gw.AuthHandler())

	api := r.Group("/api/v1")
	api.Use(gw.JWTMiddleware())
	api.Any("/users/*path", gw.UserHandler())
	api.Any("/analysis", gw.MatchHandler())
	api.Any("/chat", gw.ChatHandler())

	r.Run()
}
