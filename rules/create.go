package rules

import (
	"errors"
	"math/rand"

	"github.com/laingawbl/engine/controller/pb"
	uuid "github.com/satori/go.uuid"
)

// GameMode represents the mode the game is running in
type GameMode string

const (
	// GameModeSinglePlayer represents the game running in single player mode, this means the game will
	// run until the only snake in the game dies
	GameModeSinglePlayer GameMode = "single-player"
	// GameModeMultiPlayer represents when there is more then 1 snake in the game, this means the game will
	// run until there is zero or one snakes left alive in the game.
	GameModeMultiPlayer GameMode = "multi-player"
)

func getSnakeTimeout(req *pb.CreateRequest) int32 {
	snakeTimeout := req.SnakeTimeout
	if snakeTimeout < 1 || snakeTimeout > 5000 {
		snakeTimeout = 500
	}
	return snakeTimeout
}

// CreateInitialGame creates a new game based on the create request passed in
func CreateInitialGame(req *pb.CreateRequest) (*pb.Game, []*pb.GameFrame, error) {
	snakes, err := getSnakes(req)
	if err != nil {
		return nil, nil, err
	}
	food := []*pb.Point{}
	snakeTimeout := getSnakeTimeout(req)

	id := uuid.NewV4().String()
	game := &pb.Game{
		ID:                      id,
		Width:                   req.Width,
		Height:                  req.Height,
		Status:                  string(GameStatusStopped),
		SnakeTimeout:            snakeTimeout,
		Mode:                    string(GameModeMultiPlayer),
		MaxTurnsToNextFoodSpawn: -1,
	}

	if len(snakes) == 1 {
		game.Mode = string(GameModeSinglePlayer)
	}

	frames := []*pb.GameFrame{
		{
			Turn:   0,
			Food:   food,
			Snakes: snakes,
		},
	}

	notifyGameStart(game, frames[0])

	return game, frames, nil
}

func isTournamentBoardSize(req *pb.CreateRequest) bool {
	return isSmallBoard(req) || isMediumBoard(req) || isLargeBoard(req)
}

func isSmallBoard(req *pb.CreateRequest) bool {
	return req.Width == 7 && req.Height == 7
}

func isMediumBoard(req *pb.CreateRequest) bool {
	return req.Width == 11 && req.Height == 11
}

func isLargeBoard(req *pb.CreateRequest) bool {
	return req.Width == 19 && req.Height == 19
}

func getTournamentStartPoint(size, index int32, snakes []*pb.Snake) *pb.Point {
	if size == 7 {
		return smallStarts[index]
	} else if size == 11 {
		return mediumStarts[index]
	} else if size == 19 {
		return largeStarts[index]
	}

	return getUnoccupiedPoint(size, size, []*pb.Point{}, snakes)
}

func getSnakes(req *pb.CreateRequest) ([]*pb.Snake, error) {
	var snakes []*pb.Snake
	even := rand.Float32() < 0.5
	for index, opts := range req.Snakes {
		var startPoint *pb.Point
		if isTournamentBoardSize(req) {
			startPoint = getTournamentStartPoint(req.Width, int32(index), snakes)
		} else {
			if even {
				startPoint = getUnoccupiedPointEven(req.Width, req.Height, []*pb.Point{}, snakes)
			} else {
				startPoint = getUnoccupiedPointOdd(req.Width, req.Height, []*pb.Point{}, snakes)
			}
		}
		if startPoint == nil {
			return nil, errors.New("no unoccupied spots left for new snake")
		}
		snake := &pb.Snake{
			ID:       opts.ID,
			Name:     opts.Name,
			URL:      opts.URL,
			Health:   1000,
			HeadType: opts.HeadType,
			TailType: opts.TailType,
			Body: []*pb.Point{
				startPoint,
				startPoint.Clone(),
				startPoint.Clone(),
				startPoint.Clone(),
			},
		}
		if len(snake.ID) == 0 {
			snake.ID = uuid.NewV4().String()
		}

		for _, s := range snakes {
			if s.ID == snake.ID {
				return nil, errors.New("duplicate snake id found, create aborted")
			}
		}

		snakes = append(snakes, snake)
	}

	return snakes, nil
}
