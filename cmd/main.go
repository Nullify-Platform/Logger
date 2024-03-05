// Package main is an example application that uses the logger package
package main

import "github.com/nullify-platform/logger/pkg/logger"

func main() {
	log, err := logger.ConfigureProductionLogger("info")

	if err != nil {
		logger.Error("error configuring logger", logger.Err(err))
		panic(err)
	}
	defer log.Sync()
}
