package rules

import (
	"encoding/json"
	"time"

	"github.com/laingawbl/engine/controller/pb"
)

// SnakeUpdate bundles together a snake with a move for processing
type SnakeUpdate struct {
	Snake   *pb.Snake
	Latency time.Duration
	Move    string
	Err     error
}

func toSnakeUpdate(resp snakeResponse) *SnakeUpdate {
	if resp.err == nil {
		moveResponse := MoveResponse{}
		err := json.Unmarshal(resp.data, &moveResponse)
		if err != nil {
			return &SnakeUpdate{
				Snake:   resp.snake,
				Latency: resp.latency,
				Err:     err,
			}
		}
		return &SnakeUpdate{
			Snake:   resp.snake,
			Latency: resp.latency,
			Move:    moveResponse.Move,
		}
	}

	return &SnakeUpdate{
		Snake:   resp.snake,
		Latency: resp.latency,
		Err:     resp.err,
	}
}

// GatherSnakeMoves goes and queries each snake for the snake move
func GatherSnakeMoves(timeout time.Duration, game *pb.Game, gameFrame *pb.GameFrame) []*SnakeUpdate {
	responses := gatherAliveSnakeResponses(multiSnakeRequest{
		url:     "move",
		timeout: timeout,
		game:    game,
		frame:   gameFrame,
	})

	ret := []*SnakeUpdate{}
	for _, resp := range responses {
		ret = append(ret, toSnakeUpdate(resp))
	}
	return ret
}
