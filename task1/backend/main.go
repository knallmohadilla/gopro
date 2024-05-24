package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"github.com/rs/cors"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Game struct {
	ID string `json:"id"`
	PlayerA string `json:"playerA"`
	PlayerB string `json:"playerB"`
	ChoiceA string `json:"choiceA"`
	ChoiceB string `json:"choiceB"`
	WinsA	int `json:"winsA"`
	WinsB	int `json:"winsB"`
	GameFinished bool `json:"gameFinished"`
}

type GameChoice struct {
	ID string `json:"id"`
	Player string `json:"player"`
	GameChoice string `json:"gameChoice"`
}

type KVController struct {
	KV jetstream.KeyValue
	CTX context.Context
	MUX *http.ServeMux
}

func createGame(ctx context.Context, game Game, kv interface{}) {
	gameBytes, _ := json.Marshal(game)
	kv_:= kv.(jetstream.KeyValue)
	kv_.Put(ctx, game.ID, []byte(gameBytes))
	fmt.Printf(game.ID)

}

func updateGame(ctx context.Context, game Game, kv interface{}) {
	gameBytes, _ := json.Marshal(game)
	kv_:= kv.(jetstream.KeyValue)
	kv_.Put(ctx, game.ID, []byte(gameBytes))
	fmt.Printf(game.ID)
}

func getGame(ctx context.Context, gameID string, kv interface{}) Game {
	kv_:= kv.(jetstream.KeyValue)
	entry, _ := kv_.Get(ctx, gameID)
	var game Game
	json.Unmarshal(entry.Value(), &game)
	return Game(game)
}

func deleteGame(ctx context.Context,gameID string, kv interface{}) {
	kv_:= kv.(jetstream.KeyValue)
	kv_.Delete(ctx, gameID)	
}

func (c *KVController) GetGame(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetGame")
	idStr := r.PathValue("id")
	fmt.Println(idStr)
	task, err := c.KV.Get(c.CTX, idStr)
	var gameview Game
	json.Unmarshal(task.Value(), &gameview)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(gameview); err != nil {
		http.Error(w, "Error encountered", http.StatusInternalServerError)
	}
}

func (c *KVController) setChoice(w http.ResponseWriter, r *http.Request) {
	var gameChoice GameChoice
	if err := json.NewDecoder(r.Body).Decode(&gameChoice); err != nil {
		http.Error(w, "Invalid task data", http.StatusBadRequest)
		return
	}
	
	game:= getGame(c.CTX, gameChoice.ID, c.KV)
	if(!game.GameFinished) {
		if(gameChoice.Player != "" && gameChoice.Player == game.PlayerA){
			game.ChoiceA=gameChoice.GameChoice
		} else if (gameChoice.Player != "" && gameChoice.Player == game.PlayerB){
			game.ChoiceB=gameChoice.GameChoice
		}

		if(game.ChoiceA != "" && game.ChoiceB != "" && game.PlayerA != "" && game.PlayerB != ""){
			if(game.ChoiceA == "rock" && game.ChoiceB == "scissors"){
				game.WinsA++
			} else if(game.ChoiceA == "scissors" && game.ChoiceB == "paper"){
				game.WinsA++
			} else if(game.ChoiceA == "paper" && game.ChoiceB == "rock"){
				game.WinsA++
			} else if(game.ChoiceB == "rock" && game.ChoiceA == "scissors"){
				game.WinsB++
			} else if(game.ChoiceB == "scissors" && game.ChoiceA == "paper"){
				game.WinsB++
			} else if(game.ChoiceB == "paper" && game.ChoiceA == "rock"){
				game.WinsB++
			}
			game.GameFinished = true;
		}

	} 

	updateGame(c.CTX, game, c.KV)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(game); err != nil {
		http.Error(w, "Error encountered", http.StatusInternalServerError)
	}
}

func (c *KVController) newExistingGame(w http.ResponseWriter, r *http.Request) {
	var game Game
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		http.Error(w, "Invalid task data", http.StatusBadRequest)
		return
	}
	
	newGame:= getGame(c.CTX, game.ID, c.KV)
	if(game.GameFinished) {
		newGame.GameFinished = false;
		newGame.ChoiceA = "";
		newGame.ChoiceB = "";
		updateGame(c.CTX, newGame, c.KV)
	} 

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(newGame); err != nil {
		http.Error(w, "Error encountered", http.StatusInternalServerError)
	}
}

func (c *KVController) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	c.MUX.HandleFunc(pattern, handler)
}

func main() {	
	var firstGameName = "game" + strconv.Itoa(int(time.Now().Unix()))
	var game Game = Game{ID: firstGameName, PlayerA: "", PlayerB: "", ChoiceA: "", ChoiceB: "", GameFinished: false}
	var bucket="Games"
	
	nc, err := nats.Connect("nats://nats-service:4222") // Assign error value to blank identifier _
	if err != nil {
		fmt.Println("Error connecting to NATS:", err)
		return
	}
	defer nc.Drain()

	js, _ := jetstream.New(nc)

	var kv jetstream.KeyValue
	
	ctx := context.Background()
	//defer cancel()

	kv, _ = js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: bucket,
	})

	mux := http.NewServeMux()

	kvController := KVController{
		KV: kv,
		CTX: ctx,
		MUX: mux,
	}

	mux.HandleFunc("POST /choice", kvController.setChoice)
	mux.HandleFunc("POST /restart", kvController.newExistingGame)
	mux.HandleFunc("GET /games/{id}", kvController.GetGame)

	cons, _ := js.CreateOrUpdateConsumer(ctx, fmt.Sprintf("KV_%s", kv.Bucket()), jetstream.ConsumerConfig{
		AckPolicy: jetstream.AckNonePolicy,
		Durable:   "gamemaster",
	})

	cons.Consume(func(msg jetstream.Msg) {
		msg.Ack()
		md, _ := msg.Metadata()
		fmt.Printf("%d %s", md, string(msg.Data()))
	})
	
	fmt.Print(game)

	/*createGame(ctx, game, kv)
	fmt.Print(string(getGame(ctx, firstGameName, kv).ID))
	fmt.Print(string(getGameView(ctx, firstGameName, kv).ID))
	deleteGame(ctx, "game2", kv);*/
	
	fmt.Println("Starting server on port 1234")
	handler := cors.Default().Handler(mux)
	http.ListenAndServe(":1234", handler)
}

