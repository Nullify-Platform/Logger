// package main is an example application that uses the logger package
package main

import (
	"context"

	"github.com/nullify-platform/logger/pkg/logger"
	"github.com/nullify-platform/logger/pkg/logger/tracer"
)

func main() {
	ctx := context.Background()
	ctx, err := logger.ConfigureProductionLogger(ctx, "info")
	defer logger.L(ctx).Sync()
	ctx, span := tracer.FromContext(ctx).Start(ctx, "main")
	defer span.End()

	span.AddEvent("main function started")

	if err != nil {
		logger.L(ctx).Error("error configuring logger", logger.Err(err))
		panic(err)
	}
}
