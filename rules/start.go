package rules

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
)

// SnakeMetadata contains a snake and the metadata sent in the start response
// from the snake server.
type SnakeMetadata struct {
	Snake *pb.Snake
	Color string
	Err   error
}

func toSnakeStartResponse(resp snakeResponse) SnakeMetadata {
	if resp.err == nil {
		startResponse := StartResponse{}
		err := json.Unmarshal(resp.data, &startResponse)
		if err != nil {
			return SnakeMetadata{
				Snake: resp.snake,
				Err:   err,
			}
		}
		resp.snake.HeadType = startResponse.HeadType
		resp.snake.TailType = startResponse.TailType
		return SnakeMetadata{
			Snake: resp.snake,
			Color: startResponse.Color,
		}
	}

	return SnakeMetadata{
		Snake: resp.snake,
		Err:   resp.err,
	}
}

func isValidColour(colour string) bool {
	var re = regexp.MustCompile(`^#?[a-fA-F0-9]{6}$`)
	return re.Match([]byte(colour))
}

func getEffectiveColor(meta SnakeMetadata) string {
	if meta.Err != nil || meta.Snake == nil || !isValidColour(meta.Color) {
		return nextColor()
	}
	return meta.Color
}

// notifyGameStart calls /start on every snake and then adds metadata from the
// response to the pb.Snake object.
func notifyGameStart(game *pb.Game, startState *pb.GameFrame) {
	// Be nice and give snake servers a long time to respond to /start in case
	// it's a sleeping heroku dyno or something like that.
	timeout := 5 * time.Second
	responses := gatherSnakeStartResponses(timeout, game, startState)

	for _, resp := range responses {
		resp.Snake.Color = getEffectiveColor(resp)
	}
}

func gatherSnakeStartResponses(timeout time.Duration, game *pb.Game, startState *pb.GameFrame) []SnakeMetadata {
	responses := gatherAllSnakeResponses(multiSnakeRequest{
		url:     "start",
		timeout: timeout,
		game:    game,
		frame:   startState,
	})

	ret := []SnakeMetadata{}
	for _, resp := range responses {
		ret = append(ret, toSnakeStartResponse(resp))
	}
	return ret
}
