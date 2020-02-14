package rules

import (
	"github.com/battlesnakeio/engine/controller/pb"
)

// MoveResponse the message format of the move response from a Snake API call
type MoveResponse struct {
	Move string
}

// StartResponse is the format for /start responses
type StartResponse struct {
	Color    string
	HeadType string
	TailType string
}

// SnakeRequest the message send for all snake api calls
type SnakeRequest struct {
	Game  Game  `json:"game"`
	Turn  int32 `json:"turn"`
	Board Board `json:"board"`
	You   Snake `json:"you"`
}

// Game represents the current game state
type Game struct {
	ID string `json:"id"`
}

// Board provides information about the game board
type Board struct {
	Height int32    `json:"height"`
	Width  int32    `json:"width"`
	Food   []Coords `json:"food"`
	Snakes []Snake  `json:"snakes"`
}

// Snake represents information about a snake in the game
type Snake struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Health int32    `json:"health"`
	Body   []Coords `json:"body"`
}

// Coords represents a point on the board
type Coords struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

func buildSnakeRequest(game *pb.Game, frame *pb.GameFrame, snakeID string) SnakeRequest {
	var you *pb.Snake
	for _, s := range frame.Snakes {
		if s.ID == snakeID {
			you = s
			break
		}
	}
	return SnakeRequest{
		Game: Game{ID: game.ID},
		Turn: frame.Turn,
		Board: Board{
			Height: game.Height,
			Width:  game.Width,
			Food:   convertPoints(frame.Food),
			Snakes: convertSnakes(frame.AliveSnakes()),
		},
		You: convertSnake(you),
	}
}

func convertPoints(points []*pb.Point) []Coords {
	coords := []Coords{}

	for _, p := range points {
		coords = append(coords, Coords{X: p.X, Y: p.Y})
	}

	return coords
}

func convertSnakes(pbSnakes []*pb.Snake) []Snake {
	snakes := []Snake{}

	for _, s := range pbSnakes {
		snakes = append(snakes, convertSnake(s))
	}

	return snakes
}

func convertSnake(snake *pb.Snake) Snake {
	return Snake{
		ID:     snake.ID,
		Name:   snake.Name,
		Health: snake.Health,
		Body:   convertPoints(snake.Body),
	}
}
