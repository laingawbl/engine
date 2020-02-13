package rules

import (
	"testing"
	"time"

	"github.com/laingawbl/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestGatherSnakeMoves(t *testing.T) {
	updates := make(chan *SnakeUpdate)

	gatherMoveResponses(t, "{\"move\":\"up\"}", updates)

	select {
	case update := <-updates:
		require.NoError(t, update.Err)
		require.Equal(t, "up", update.Move)
		require.NotEmpty(t, update.Latency)
	case <-time.After(250 * time.Millisecond):
		require.Fail(t, "No update received over updates channel")
	}
}

func TestGatherSnakeMovesBadJSON(t *testing.T) {
	updates := make(chan *SnakeUpdate)

	gatherMoveResponses(t, "{{", updates)

	select {
	case update := <-updates:
		require.Error(t, update.Err)
	case <-time.After(250 * time.Millisecond):
		require.Fail(t, "No update received over updates channel")
	}
}

func gatherMoveResponses(t *testing.T, json string, updates chan<- *SnakeUpdate) {
	createClient = singleEndpointMockClient(t, "http://not.a.snake.com/move", json, 200)

	go func() {
		u := GatherSnakeMoves(1*time.Second, &pb.Game{}, &pb.GameFrame{
			Snakes: []*pb.Snake{
				{
					URL: "http://not.a.snake.com",
				},
			},
		})
		if len(u) > 0 {
			updates <- u[0]
		}
	}()
}
