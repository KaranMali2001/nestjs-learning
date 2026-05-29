package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)
	code := status.Code(err)
	fmt.Printf("[gRPC] %-60s code=%s duration=%s\n", info.FullMethod, code, duration)
	return resp, err
}
