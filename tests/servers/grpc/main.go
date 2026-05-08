package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/VDHewei/xsh/tests/servers/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var port = getEnv("MOCK_GRPC_PORT", "18081")

func main() {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMockServiceServer(s, &mockServer{})
	reflection.Register(s)

	log.Printf("Mock gRPC server starting on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

type mockServer struct {
	pb.UnimplementedMockServiceServer
}

func (s *mockServer) HealthCheck(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	// Simulate failure for "fail" service
	if req.GetService() == "fail" {
		return &pb.HealthResponse{
			Healthy:   false,
			Message:   "service unhealthy",
			Timestamp: time.Now().Unix(),
		}, nil
	}
	return &pb.HealthResponse{
		Healthy:   true,
		Message:   fmt.Sprintf("service %s is healthy", req.GetService()),
		Timestamp: time.Now().Unix(),
	}, nil
}

func (s *mockServer) Execute(ctx context.Context, req *pb.ExecuteRequest) (*pb.ExecuteResponse, error) {
	// Simulate failure for "fail" method
	if req.GetMethod() == "fail" {
		return &pb.ExecuteResponse{
			Success: false,
			Error:   "execution failed",
		}, nil
	}
	return &pb.ExecuteResponse{
		Success: true,
		Output:  fmt.Sprintf("executed %s with payload: %s", req.GetMethod(), req.GetPayload()),
	}, nil
}

func (s *mockServer) StreamExecute(req *pb.ExecuteRequest, stream pb.MockService_StreamExecuteServer) error {
	for i := 0; i < 3; i++ {
		resp := &pb.ExecuteResponse{
			Success: true,
			Output:  fmt.Sprintf("stream chunk %d for %s", i+1, req.GetMethod()),
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (s *mockServer) Timeout(ctx context.Context, req *pb.TimeoutRequest) (*pb.TimeoutResponse, error) {
	delay := time.Duration(req.GetDelaySeconds()) * time.Second
	select {
	case <-time.After(delay):
		return &pb.TimeoutResponse{Message: "completed after delay"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *mockServer) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{
		Method:  req.GetMethod(),
		Payload: req.GetPayload(),
		Headers: req.GetHeaders(),
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
