package main

import (
    "log"
    "net/http"
    "os"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

type GameState struct {
    Board         [9]string
    CurrentPlayer string
    Players       map[string]*websocket.Conn
    PlayersReady  int
    Winner        string
    GameOver      bool
    Mutex         sync.Mutex
}

var (
    games     = make(map[string]*GameState)
    gamesLock sync.Mutex
    upgrader  = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        HandshakeTimeout: 10 * time.Second,
    }
)

func checkWinner(board [9]string) string {
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

func resetGame(game *GameState) {
    game.Board = [9]string{}
    game.CurrentPlayer = "O"
    game.Winner = ""
    game.GameOver = false
}

func handleGame(c *gin.Context) {
    gameId := c.Param("gameId")
    c.HTML(http.StatusOK, "game.html", gin.H{
        "gameId": gameId,
    })
}

func handleWebSocket(c *gin.Context) {
    gameId := c.Param("gameId")

    // Configure WebSocket connection with timeout
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("Failed to upgrade connection: %v", err)
        return
    }
    defer conn.Close()

    // Set read/write deadlines
    conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

    gamesLock.Lock()
    game, exists := games[gameId]
    if !exists {
        game = &GameState{
            Players:       make(map[string]*websocket.Conn),
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
        "type":          "init",
        "symbol":        playerSymbol,
        "board":         game.Board,
        "currentPlayer": game.CurrentPlayer,
        "playersReady":  game.PlayersReady,
        "gameOver":      game.GameOver,
        "winner":        game.Winner,
    })

    // Keep-alive handler
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                    return
                }
            }
        }
    }()

    // Main game loop
    for {
        var message struct {
            Type     string `json:"type"`
            Position int    `json:"position"`
        }

        err := conn.ReadJSON(&message)
        if err != nil {
            log.Printf("Error reading message: %v", err)
            break
        }

        game.Mutex.Lock()

        if message.Type == "move" && !game.GameOver {
            if game.CurrentPlayer == playerSymbol &&
                message.Position >= 0 &&
                message.Position < 9 &&
                game.Board[message.Position] == "" {

                game.Board[message.Position] = playerSymbol

                winner := checkWinner(game.Board)
                if winner != "" {
                    game.GameOver = true
                    game.Winner = winner
                } else {
                    if playerSymbol == "O" {
                        game.CurrentPlayer = "X"
                    } else {
                        game.CurrentPlayer = "O"
                    }
                }

                // Broadcast updated state with error handling
                for _, playerConn := range game.Players {
                    err := playerConn.WriteJSON(gin.H{
                        "type":          "update",
                        "board":         game.Board,
                        "currentPlayer": game.CurrentPlayer,
                        "gameOver":      game.GameOver,
                        "winner":        game.Winner,
                    })
                    if err != nil {
                        log.Printf("Error broadcasting state: %v", err)
                    }
                }
            }
        } else if message.Type == "restart" && game.GameOver {
            resetGame(game)
            for _, playerConn := range game.Players {
                err := playerConn.WriteJSON(gin.H{
                    "type":          "update",
                    "board":         game.Board,
                    "currentPlayer": game.CurrentPlayer,
                    "gameOver":      game.GameOver,
                    "winner":        game.Winner,
                })
                if err != nil {
                    log.Printf("Error broadcasting reset: %v", err)
                }
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

func main() {
    // Set Gin mode based on environment
    ginMode := os.Getenv("GIN_MODE")
    if ginMode != "" {
        gin.SetMode(ginMode)
    }

    r := gin.Default()

    // Enhanced CORS middleware for Render
    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
        c.Writer.Header().Set("Access-Control-Max-Age", "3600")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    })

    // Serve static files
    r.Static("/static", "./static")
    r.LoadHTMLGlob("templates/*")

    // Routes
    r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", nil)
    })

    r.GET("/game/:gameId", handleGame)
    r.GET("/ws/:gameId", handleWebSocket)

    // Get port from environment variable
    port := os.Getenv("PORT")
    if port == "" {
        port = "10000" // Render default port
    }

    // Start server with improved logging
    log.Printf("Server starting on port %s", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatal("Failed to start server: ", err)
    }
}