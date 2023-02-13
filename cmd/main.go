package main

import "github.com/Nullify-Platform/Logger/pkg/logger"

func main() {
	log, err := logger.ConfigureProductionLogger("info")

	if err != nil {
		logger.Error("error configuring logger", logger.Err(err))
		panic(err)
	}
	defer log.Sync()
}
