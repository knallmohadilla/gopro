package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "github.com/knallmohadilla/gopro/task2/rps"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedGameServiceServer
}



func (s *server) CreateGame(ctx context.Context, in *pb.GameRequest) (*pb.GameResponse, error) {
	log.Printf("Received: %v", in.GetGame())
	return &pb.GameResponse{Game: &pb.Game{Id: "1", Name: "game1"}}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGameServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}