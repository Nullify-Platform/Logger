// package main is an example application that uses the logger package
package main

import (
	"context"

	"github.com/nullify-platform/logger/pkg/logger"
)

func main() {
	ctx := context.Background()
	ctx, span, err := logger.ConfigureProductionLogger(ctx, "main", "info")
	defer span.End()
	defer logger.L(ctx).Sync()

	span.AddEvent("main function started")

	if err != nil {
		logger.L(ctx).Error("error configuring logger", logger.Err(err))
		panic(err)
	}
}
