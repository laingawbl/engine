package rules

import (
	"errors"
	"testing"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestStartSnakes(t *testing.T) {
	resetPalette(defaultColors)
	snake := getSnakeAfterStart(t, "{\"color\":\"#ff0000\"}", 200)
	require.Equal(t, "#ff0000", snake.Color)
	require.Nil(t, snake.Death, "Snake should not be dead")
}

func TestStartSnakesMissingColor(t *testing.T) {
	resetPalette([]string{"red", "green", "blue"})
	snake := getSnakeAfterStart(t, "{}", 200)
	require.Equal(t, "red", snake.Color)
	require.Nil(t, snake.Death, "Snake should not be dead")
}

func TestStartSnakesMissingEndpoint(t *testing.T) {
	resetPalette([]string{"red", "green", "blue"})
	snake := getSnakeAfterStart(t, "{}", 404)
	require.Equal(t, "red", snake.Color)
	require.Nil(t, snake.Death, "Snake should not be dead")
}

func TestStartSnakesMissingServer(t *testing.T) {
	resetPalette([]string{"red", "green", "blue"})
	snake := getSnakeAfterMissingServer(t)
	require.Equal(t, "red", snake.Color)
	require.Nil(t, snake.Death, "Snake should not be dead")
}

func getSnakeAfterStart(t *testing.T, json string, statusCode int) *pb.Snake {
	url := "http://good-server/start"
	createClient = singleEndpointMockClient(t, url, json, statusCode)

	snake := &pb.Snake{
		URL: "http://good-server",
	}

	notifyGameStart(&pb.Game{}, &pb.GameFrame{
		Snakes: []*pb.Snake{snake},
	})

	return snake
}

func getSnakeAfterMissingServer(t *testing.T) *pb.Snake {
	createClient = func(time.Duration) httpClient {
		return mockHTTPClient{
			err: errors.New("fail"),
		}
	}

	snake := &pb.Snake{
		URL: "http://dead-server",
	}

	notifyGameStart(&pb.Game{}, &pb.GameFrame{
		Snakes: []*pb.Snake{snake},
	})

	return snake
}
