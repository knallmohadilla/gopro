package main

import (
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/knallmohadilla/gopro/task2/rps"
)


var (
	addr = flag.String("addr", "localhost:50052", "the address to connect to")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGameServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.CreateGame(ctx, &pb.CreateGameRequest{Player: &pb.Player{Name: "Player A", Choice: "rock"}})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf(r.Game.GetId(), r.Game.GetPlayerA())
}
