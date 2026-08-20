package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/mlplabs/lynway-wms/internal/app"
	"github.com/mlplabs/lynway-wms/internal/config"
)

// @title Lynway WMS API
// @version 1.0
func main() {
	_ = godotenv.Load(".env")

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("main. Ошибка при получении конфигурации: %s", err)
	}

	app.Run(cfg)
}
