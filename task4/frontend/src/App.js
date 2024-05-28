import logo from './logo.svg';
import React, {useEffect, useState, useReducer} from 'react';
import {Buffer} from 'buffer';
import './App.css';
import {
  connect,
  StringCodec,
  JSONCodec
} from "nats.ws";
//import {useNats, useNatsSubscription} from "./useNats.ts"



function App() {
  const style = 
    {
      margin: 5 + 'px',
      width: 70 + 'px',
      display: "inline-block",
      fontSize: 14 + 'px',
    }
    const inputstyle = 
    {
      margin: 5 + 'px',
      width: 105 + 'px',
      display: "inline-block",
      fontSize: 14 + 'px',
      marginTop:10 + 'px',
      marginBottom:20 + 'px'
    }
    const createstyle = 
    {
      margin: 5 + 'px',
      width: 110 + 'px',
      display: "inline-block",
      fontSize: 14 + 'px',
    }

  const highstyle = 
    {
      margin: 5 + 'px',
      width: 70 + 'px',
      display: "inline-block",
      background: "green",
      color: "white",
      fontSize: 14 + 'px',
    }
  
  const players = [
    "Odin",
    "Loki",
    "Thor",
    "Freya",
    "Frigg",
    "Balder",
    "Heimdall",
    "Tyr",
    "Freyr",
    "Njord",
  ]
  const sc = new StringCodec();
  const jc = JSONCodec();
  const [nats, setNats] = useState(undefined);
  const [sub, setSub] = useState(undefined);
  const [viewMessage, setViewMessage] = useState("");
  const [loadKV, setLoadKV] = useState(undefined);
  const [games, setGames] = useState([]); 
  const [kv, setKV] = useState(undefined);
  const [input, setInput] = useState(players[Math.floor(Math.random() * players.length)]);
  const [userChoice, setUserChoice] = useState('');
  const ip = "http://" + window.location.href.split("//")[1].split("/")[0].split(':')[0];
  
  useEffect(() => {
    if(ip) {
      (async () => {
          const nc = await connect({
            servers: [ip + ":30090"], // ip = http://192.168.49.2
          })
          
          console.log("connected to NATS")
          const js = nc.jetstream();
          const kv = await js.views.kv("Games");

          const watcher = await kv.watch({
            key: ">",
          });

          setKV(kv);

          for await (const e of watcher) {
            try {
              const res = JSON.parse(String(Buffer.from(e.value).toString('utf8')));
              setGames(prev => [...prev.filter(game => game.id !== res.id), res] );
            } catch (error) {
              console.log("error");
            }
          }
        })();
    }
  
    return () => {
      nats?.drain();
      console.log("closed NATS connection")
    }
  }, [])
  
  const loadGames = (e) => {
    setLoadKV(true);
  }

  const removeGame = async (e, id) => {
    if(kv) {
      setGames(prev => prev.filter(game => game.id !== id));
      await kv.purge(id);
    } else {
      console.log("nokv");
    }
  }

  const joinGame = async (e, id) => {
    if(kv) {
      for(let i = 0; i < games.length; i++) {
        if(games[i].id === id) {
          if(games[i].playerA === "") {
            if(games[i].playerB !== input) {
              games[i].playerA = input;
              setGames(prev => prev.filter(game => game.id !== games[i].id))
              await kv.put(id, jc.encode(games[i]));
              
            } else {
              console.log("You are already in this game");
            }
          } else {
            if(games[i].playerA !== input) {
              games[i].playerB = input;
              setGames(prev => prev.filter(game => game.id !== games[i].id))
              await kv.put(id, jc.encode(games[i]));
            } else {
              console.log("You are already in this game");
            }
          }
        }
      }
    } else {
      console.log("nokv");
    }
  }

  const setChoice = async (e, id, player, choice) => {
    try {
      if (ip) {
        const response = await fetch(ip + ':30034/choice', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            id: id,
            player: player,
            gameChoice: choice
          })
        })
        let data = await response.json();
        setGames(prev => [...prev.filter(game => game.id !== id), data])
        setUserChoice(choice)
        if(data.currentWinner === player) {
          alert("You won, hit restart if you want to play again!")
        } else if(data.currentWinner === "draw") {
          alert("It's a draw, hit restart if you want to play again!")
        } else {
          alert("You lost, hit restart if you want to play again!")
        }
      }
    } catch (error) {
      console.log("error");
    }
  }

  const restartGame = async (e, id) => {
    try {
      if (ip) {
        console.log(games.filter(game => game.id === id))
        const response = await fetch(ip + ':30034/restart', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(games.filter(game => game.id === id)[0])
        })
        let data = await response.json();
        console.log(data)
        setGames(prev => [...prev.filter(game => game.id !== id), data])
        }
    } catch (error) {
      console.log("error");
    }
  }

  const createGame = async (e) => {
    try {
      const id="game" + new Date().getUTCSeconds()
      await kv.put(id, jc.encode({
        "id":id,
        "playerA":"",
        "playerB":"",
        "choiceA":"",
        "choiceB":"",
        "winsA":0,
        "winsB":0,
        "gameFinished":false,
        "currentWinner":""
      }));
    } catch (error) {
      console.log("error");
    }
  }

  return (
    <div className="App">
      <h3>Rock, Paper, Scissors - Create or Join a Game, choose your Gesture, win or loose & repeat!</h3>
      <span style={style}>Username</span>
      {
        <input style={inputstyle} value={input} type='text' onInput={e => setInput(e.target.value)}/>
      }
      {
        <input style={createstyle}  type='button' onClick={(e) => createGame(e)} value="+ Add Game"/>
      }
      {
        games.map((game, index) => (
          <div key={index}>
            <span style={style}>{game.playerA === "" ? "none" : game.playerA + "(" + game.winsA + ")"}</span>vs.<span style={style}>{game.playerB === "" ? "none" : game.playerB + "(" + game.winsB + ")"}</span>
            <span style={style}>{
              game.playerA === "" || game.playerB === "" && game.playerB !== input && game.playerA !== input? <input type='button' onClick={(e) => joinGame(e, game.id)} value="Join"/> : 
              game.gameFinished ?
              <input style={style} type='button' onClick={(e) => restartGame(e, game.id)} value="restart"/>
              : <></>
            }</span>
            {
              <input style={
                (game.playerA === input && game.choiceA === "rock") 
                || (game.playerB === input && game.choiceB === "rock")  
                ? highstyle 
                : style
              }  type='button' onClick={(e) => setChoice(e, game.id, input, "rock")} value="rock"/>
            }
            {
              <input style={
                (game.playerA === input && game.choiceA === "paper") 
                || (game.playerB === input && game.choiceB === "paper")  
                ? highstyle 
                : style
              }  type='button' onClick={(e) => setChoice(e, game.id, input, "paper")} value="paper"/>
            }
            {
              <input style={
                (game.playerA === input && game.choiceA === "scissors") 
                || (game.playerB === input && game.choiceB === "scissors")  
                ? highstyle 
                : style
              }  type='button' onClick={(e) => setChoice(e, game.id, input, "scissors")} value="scissor"/>
            }
            {
              <input style={style} type='button' onClick={(e) => removeGame(e, game.id)} value="delete -"/>
            }
          </div>
        ))
      }
    </div>
  );
}

export default App;
