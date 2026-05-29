package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKey string

const AuthKey ctxKey = "auth"

func AuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if _, ok := publicMethods[info.FullMethod]; ok {
		return handler(ctx, req)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		fmt.Println("NO metadata found")
		return nil, status.Error(codes.Unauthenticated, "missing Metadata ")

	}
	authValues := md.Get("auth")
	if len(authValues) == 0 {
		fmt.Println("NO metadata found in auth")
		return nil, status.Error(codes.Unauthenticated, "missing auth header")
	}

	ctx = context.WithValue(ctx, AuthKey, authValues[0])
	return handler(ctx, req)
}
