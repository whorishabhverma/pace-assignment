// main.go
package main

import (
    "log"
    "net/http"
    "sync"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

type GameState struct {
    Board        [9]string
    CurrentPlayer string
    Players      map[string]*websocket.Conn
    PlayersReady int
    Winner       string
    GameOver     bool
    Mutex        sync.Mutex
}

var (
    games = make(map[string]*GameState)
    gamesLock sync.Mutex
    upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
    }
)

func checkWinner(board [9]string) string {
    // Winning combinations
    winCombos := [][3]int{
        {0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // Rows
        {0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // Columns
        {0, 4, 8}, {2, 4, 6},            // Diagonals
    }

    for _, combo := range winCombos {
        if board[combo[0]] != "" &&
           board[combo[0]] == board[combo[1]] &&
           board[combo[1]] == board[combo[2]] {
            return board[combo[0]]
        }
    }

    // Check for draw
    isDraw := true
    for _, cell := range board {
        if cell == "" {
            isDraw = false
            break
        }
    }
    if isDraw {
        return "draw"
    }

    return ""
}

func main() {
    r := gin.Default()
    r.Static("/static", "./static")
    r.LoadHTMLGlob("templates/*")
    
    r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", nil)
    })
    
    r.GET("/game/:gameId", handleGame)
    r.GET("/ws/:gameId", handleWebSocket)
    
    r.Run(":8080")
}

func handleGame(c *gin.Context) {
    gameId := c.Param("gameId")
    c.HTML(http.StatusOK, "game.html", gin.H{
        "gameId": gameId,
    })
}

func resetGame(game *GameState) {
    game.Board = [9]string{}
    game.CurrentPlayer = "O"
    game.Winner = ""
    game.GameOver = false
}

func handleWebSocket(c *gin.Context) {
    gameId := c.Param("gameId")
    
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("Failed to upgrade connection: %v", err)
        return
    }
    defer conn.Close()
    
    gamesLock.Lock()
    game, exists := games[gameId]
    if !exists {
        game = &GameState{
            Players: make(map[string]*websocket.Conn),
            CurrentPlayer: "O",
        }
        games[gameId] = game
    }
    
    playerSymbol := "O"
    if len(game.Players) == 1 {
        playerSymbol = "X"
    }
    
    game.Players[playerSymbol] = conn
    game.PlayersReady++
    gamesLock.Unlock()
    
    // Send initial game state
    conn.WriteJSON(gin.H{
        "type": "init",
        "symbol": playerSymbol,
        "board": game.Board,
        "currentPlayer": game.CurrentPlayer,
        "playersReady": game.PlayersReady,
        "gameOver": game.GameOver,
        "winner": game.Winner,
    })
    
    // Main game loop
    for {
        var message struct {
            Type string `json:"type"`
            Position int `json:"position"`
        }
        
        err := conn.ReadJSON(&message)
        if err != nil {
            log.Printf("Error reading message: %v", err)
            break
        }
        
        game.Mutex.Lock()
        
        if message.Type == "move" && !game.GameOver {
            // Validate move
            if game.CurrentPlayer == playerSymbol && 
               message.Position >= 0 && 
               message.Position < 9 && 
               game.Board[message.Position] == "" {
                
                // Make move
                game.Board[message.Position] = playerSymbol
                
                // Check for winner
                winner := checkWinner(game.Board)
                if winner != "" {
                    game.GameOver = true
                    game.Winner = winner
                } else {
                    // Switch turns
                    if playerSymbol == "O" {
                        game.CurrentPlayer = "X"
                    } else {
                        game.CurrentPlayer = "O"
                    }
                }
                
                // Broadcast updated state
                for _, conn := range game.Players {
                    conn.WriteJSON(gin.H{
                        "type": "update",
                        "board": game.Board,
                        "currentPlayer": game.CurrentPlayer,
                        "gameOver": game.GameOver,
                        "winner": game.Winner,
                    })
                }
            }
        } else if message.Type == "restart" && game.GameOver {
            resetGame(game)
            // Broadcast reset state
            for _, conn := range game.Players {
                conn.WriteJSON(gin.H{
                    "type": "update",
                    "board": game.Board,
                    "currentPlayer": game.CurrentPlayer,
                    "gameOver": game.GameOver,
                    "winner": game.Winner,
                })
            }
        }
        
        game.Mutex.Unlock()
    }
    
    // Clean up when player disconnects
    game.Mutex.Lock()
    delete(game.Players, playerSymbol)
    game.PlayersReady--
    if game.PlayersReady == 0 {
        gamesLock.Lock()
        delete(games, gameId)
        gamesLock.Unlock()
    }
    game.Mutex.Unlock()
}