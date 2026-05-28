package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	pb "github.com/karanmali5599/go-ts-grpc/server/gen/pipeline/v1"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedPipelineServiceServer
}

func (s *Server) CreatePipelineAndJobs(ctx context.Context, req *pb.CreatePipelineAndJobsRequest) (*pb.CreatePipelineAndJobsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		fmt.Println("NO metadata found")
		return nil, status.Error(codes.InvalidArgument, "missing Metadata ")

	}
	if len(md.Get("auth")) == 0 {
		fmt.Println("NO metadata found in auth")
		return nil, status.Error(codes.InvalidArgument, "missing auth header")
	}
	fmt.Println("CONTEXT", md)
	fmt.Println("inside the CreatePipelineAndJobs", req.Pipeline)
	return &pb.CreatePipelineAndJobsResponse{Pipeline: req.Pipeline}, nil
}
func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to get listener %v", err)

	}
	var grpcOpts []grpc.ServerOption
	srv := grpc.NewServer(grpcOpts...)
	pb.RegisterPipelineServiceServer(srv, &Server{})
	reflection.Register(srv)

	fmt.Println("Starting server on port 8080")
	if err := srv.Serve(listener); err != nil {
		log.Fatalf("Stopped Server %v", err)
	}
}
