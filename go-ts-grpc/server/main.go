package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"buf.build/go/protovalidate"
	pb "github.com/karanmali5599/go-ts-grpc/server/gen/pipeline/v1"
	protovalidateinterceptor "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedPipelineServiceServer
}
type HealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

var publicMethods = map[string]bool{
	"/grpc.health.v1.Health/Check": true,
	"/grpc.health.v1.Health/Watch": true,
}

func (s *Server) CreatePipelineAndJobs(ctx context.Context, req *pb.CreatePipelineAndJobsRequest) (*pb.CreatePipelineAndJobsResponse, error) {

	fmt.Println("inside the CreatePipelineAndJobs", req.Pipeline)
	if req.Pipeline == nil {
		panic("TEST PANIC")
	}
	authToken, ok := ctx.Value(AuthKey).(string)
	if !ok {
		fmt.Println("auth token missing from ctx")
	}
	fmt.Println("THE AUTH TOKEN is", authToken)
	return &pb.CreatePipelineAndJobsResponse{Pipeline: req.Pipeline}, nil
}

func (h *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	fmt.Println("Checking custom health end point")
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}
func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to get listener %v", err)

	}
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("Failed to create protovalidate validator %v", err)
	}

	// this is creating the server which owns the socket
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			LoggingInterceptor, // outermost — sees every request including auth failures
			AuthInterceptor,
			protovalidateinterceptor.UnaryServerInterceptor(validator), // validates request payload after auth, before handler
			RecoveryInterceptor, // innermost — converts handler panics to codes.Internal before outer interceptors see them
		),
	)

	// implementing the health check service which is provided by gRPC
	hsrv := &HealthServer{}
	// registering the services to server

	// health service by grpc
	grpc_health_v1.RegisterHealthServer(srv, hsrv)
	// registering the pipeline service
	pb.RegisterPipelineServiceServer(srv, &Server{})
	reflection.Register(srv)

	fmt.Println("Starting server on port 8080")
	if err := srv.Serve(listener); err != nil {
		log.Fatalf("Stopped Server %v", err)
	}
}
