package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/knallmohadilla/gopro/task2/rps"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50052, "The server port")
)

type server struct {
	pb.UnimplementedGameServiceServer
}



func (s *server) GameService(ctx context.Context, in *pb.CreateGameRequest) (*pb.GameResponse, error) {
	log.Printf("Received: %v", in.GetPlayer())
	gameID := fmt.Sprintf("game-%d", time.Now().Unix())
	player := in.Player.Name;
	choice:= in.Player.Choice;
	if player == "" {
		player = "Player A"
	}
	
	var choiceA string
	if in.Player.Choice != "" {
		choiceA = choice
	} else {
		choiceA = ""
	}
	
	return &pb.GameResponse{Game: &pb.Game{
		Id:            gameID,
		PlayerA:       player,
		PlayerB:       "",
		ChoiceA:       choiceA,
		ChoiceB:       "",
		WinsA:         0,
		WinsB:         0,
		GameFinished:  false,
		CurrentWinner: "",
	}}, nil
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