package rules

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/laingawbl/engine/controller/pb"
)

func TestCheckForGameOver_SinglePlayer(t *testing.T) {

	gameFrame := &pb.GameFrame{
		Snakes: []*pb.Snake{
			{Death: &pb.Death{}},
		},
	}
	res := CheckForGameOver(GameModeSinglePlayer, gameFrame)
	require.True(t, res)

	gameFrame.Snakes[0].Death = nil
	res = CheckForGameOver(GameModeSinglePlayer, gameFrame)
	require.False(t, res)
}
