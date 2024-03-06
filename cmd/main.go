// Package main is an example application that uses the logger package
package main

import (
	"context"
	"errors"
	"github.com/nullify-platform/logger/pkg/logger"
	"github.com/nullify-platform/logger/pkg/logger/tracer"
)

func main() {
	ctx := context.Background()
	ctx, err := logger.ConfigureProductionLogger(ctx, "info")
	defer logger.FromContext(ctx).Sync()
	ctx, span := tracer.FromContext(ctx).Start(ctx, "main")
	defer span.End()

	span.AddEvent("main function started")

	if err != nil {
		logger.FromContext(ctx).Error("error configuring logger", logger.Err(err))
		panic(err)
	}

	anotherFunctionRenamed(ctx)
}

func anotherFunctionRenamed(ctx context.Context) {
	ctx, span := tracer.FromContext(ctx).Start(ctx, "extended feature")
	defer span.End()

	logger.FromContext(ctx).Info("another function started")
	logger.FromContext(ctx).Error("something terbl happened", logger.Err(errors.New("test error 2 in dev")))
}
