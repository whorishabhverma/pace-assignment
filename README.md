# Real-Time Multiplayer Tic Tac Toe

A real-time multiplayer Tic Tac Toe game built with Go, WebSocket, and vanilla JavaScript. Players can create game rooms and share links with friends to play together in real-time.

## 🎮 Demo

Visit [your-deployed-url.onrender.com](https://tic-tac-toe-pi9m.onrender.com)  to try the game.

## ✨ Features

- Real-time multiplayer gameplay using WebSockets
- Unique game rooms with shareable links
- Clean and responsive UI
- No database required - all game state managed in memory
- Cross-browser compatible
- Mobile-friendly design

## 🛠️ Technology Stack

- **Backend**: Go (Gin Framework)
- **Frontend**: HTML, CSS, JavaScript (Vanilla)
- **Real-time Communication**: WebSocket (Gorilla WebSocket)
- **Deployment**: Render

## 🚀 Getting Started

### Prerequisites

- Go 1.16 or higher
- Git

### Local Development

1. Clone the repository
```bash
git clone https://github.com/your-username/tic-tac-toe.git
cd tic-tac-toe
```

2. Install dependencies
```bash
go mod tidy
```

3. Run the server
```bash
go run main.go
```

4. Open your browser and visit `http://localhost:10000`

## 📝 How to Play

1. Visit the homepage and click "Create New Game"
2. Share the generated game link with your friend
3. Wait for your friend to join
4. Play turns alternately until someone wins or the game draws
5. Click "Play Again" to start a new game

## 🎯 Game Rules

- The game is played on a 3x3 grid
- Players take turns putting their marks (O or X) in empty squares
- The first player to get 3 of their marks in a row (vertically, horizontally, or diagonally) wins
- When all 9 squares are full, the game is a draw

## 🔧 Project Structure

```
tic-tac-toe/
├── main.go           # Main server file
├── static/
│   └── css/
│       ├── styles.css    # Homepage styles
│       └── game.css      # Game page styles
└── templates/
    ├── index.html    # Homepage template
    └── game.html     # Game page template
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔮 Future Improvements

- [ ] Add game history
- [ ] Implement user authentication
- [ ] Add game timer
- [ ] Save game statistics
- [ ] Add sound effects
- [ ] Support custom board sizes
- [ ] Add chat functionality
- [ ] Implement game replay feature

## 📞 Contact

Your Name - [your-email@example.com](mailto:rishabh63885@example.com)

Project Link: [https://github.com/your-username/tic-tac-toe](https://github.com/whorishabhverma/pace-assignment)
