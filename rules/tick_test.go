package rules

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/laingawbl/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestGetUnoccupiedPointEven(t *testing.T) {

	unoccupiedPoint := getUnoccupiedPointEven(2, 2,
		[]*pb.Point{},
		[]*pb.Snake{})
	require.True(t, (unoccupiedPoint.X+unoccupiedPoint.Y)%2 == 0, "Point coordinates should sum to an even number %o ", unoccupiedPoint)
}

func TestGetUnoccupiedPointOdd(t *testing.T) {
	unoccupiedPoint := getUnoccupiedPointOdd(2, 2,
		[]*pb.Point{{X: 0, Y: 1}},
		[]*pb.Snake{})
	require.True(t, (unoccupiedPoint.X+unoccupiedPoint.Y)%2 == 1, "Point coordinates should sum to an odd number %o ", unoccupiedPoint)
}

func TestGetUnoccupiedPointWithFullBoard(t *testing.T) {
	unoccupiedPoint := getUnoccupiedPoint(2, 2,
		[]*pb.Point{{X: 0, Y: 0}},
		[]*pb.Snake{
			{
				Body: []*pb.Point{
					{X: 0, Y: 1},
					{X: 1, Y: 1},
					{X: 1, Y: 0},
				},
			},
		})
	require.True(t, unoccupiedPoint.Equal(nil))
}

func TestGetUnoccupiedPointsWithEmptySpots(t *testing.T) {
	unoccupiedPoints := getUnoccupiedPoints(2, 2,
		[]*pb.Point{{X: 0, Y: 0}},
		[]*pb.Snake{
			{
				Body: []*pb.Point{
					{X: 0, Y: 1},
				},
			},
		})

	require.Len(t, unoccupiedPoints, 2)
	require.True(t, unoccupiedPoints[0].Equal(&pb.Point{X: 1, Y: 0}))
	require.True(t, unoccupiedPoints[1].Equal(&pb.Point{X: 1, Y: 1}))
}

func TestGetUniqOccupiedPoints(t *testing.T) {
	unoccupiedPoints := getUniqOccupiedPoints(
		[]*pb.Point{
			{X: 0, Y: 0},
		},
		[]*pb.Snake{
			{
				Body: []*pb.Point{
					{X: 0, Y: 1},
					{X: 1, Y: 1},
					{X: 1, Y: 1},
					{X: 1, Y: 0},
				},
			},
		})

	require.Len(t, unoccupiedPoints, 4)
}

func TestGameTickUpdatesTurnCounter(t *testing.T) {
	gt, err := GameTick(commonGame, &pb.GameFrame{Turn: 5})
	require.NoError(t, err)
	require.Equal(t, int32(6), gt.Turn)
}

func TestGameTickUpdatesSnake(t *testing.T) {
	snake := &pb.Snake{
		Health: 67,
		Body: []*pb.Point{
			{X: 1, Y: 1},
			{X: 1, Y: 2},
			{X: 1, Y: 3},
		},
	}
	game := &pb.Game{
		Width:  20,
		Height: 20,
	}
	gt, err := GameTick(game, &pb.GameFrame{
		Turn: 5,
		Snakes: []*pb.Snake{
			snake,
		},
	})
	require.NoError(t, err)
	require.Len(t, gt.Snakes, 1)
	snake = gt.Snakes[0]
	require.Equal(t, int32(66), snake.Health)
	require.Len(t, snake.Body, 3)
	require.Equal(t, &pb.Point{X: 1, Y: 0}, snake.Body[0])
	require.Equal(t, &pb.Point{X: 1, Y: 1}, snake.Body[1])
	require.Equal(t, &pb.Point{X: 1, Y: 2}, snake.Body[2])
}

var commonGame = &pb.Game{
	Width:  20,
	Height: 20,
}
var lastFrame = &pb.GameFrame{
	Turn:   5,
	Snakes: []*pb.Snake{},
	Food: []*pb.Point{
		{X: 1, Y: 0},
	},
}

func TestGameFrameSnakeEats(t *testing.T) {
	snake := &pb.Snake{
		Health: 67,
		Body: []*pb.Point{
			{X: 1, Y: 1},
			{X: 1, Y: 2},
			{X: 1, Y: 3},
		},
	}

	lastFrame.Snakes = []*pb.Snake{snake}

	gt, err := GameTick(commonGame, lastFrame)
	require.NoError(t, err)
	require.Len(t, gt.Snakes, 1)
	snake = gt.Snakes[0]
	require.Equal(t, int32(66), snake.Health)
	require.Len(t, snake.Body, 3)
}

func TestGameTickDeadSnakeDoNotUpdate(t *testing.T) {
	snake := &pb.Snake{
		Health: 87,
		Body: []*pb.Point{
			{X: 1, Y: 1},
			{X: 1, Y: 2},
			{X: 1, Y: 3},
		},
		Death: &pb.Death{
			Turn:  4,
			Cause: DeathCauseSnakeCollision,
		},
	}

	lastFrame.Snakes = []*pb.Snake{snake}

	gt, err := GameTick(commonGame, lastFrame)
	require.NoError(t, err)
	require.Len(t, gt.Snakes, 1)
	snake = gt.Snakes[0]
	require.Equal(t, int32(87), snake.Health)
	require.Len(t, snake.Body, 3)
	require.Equal(t, &pb.Point{X: 1, Y: 1}, snake.Body[0])
	require.Equal(t, &pb.Point{X: 1, Y: 2}, snake.Body[1])
	require.Equal(t, &pb.Point{X: 1, Y: 3}, snake.Body[2])
}

func TestGameTickUpdatesDeath(t *testing.T) {
	snake := &pb.Snake{
		Health: 0,
		Body: []*pb.Point{
			{X: 3, Y: 1},
			{X: 3, Y: 2},
			{X: 3, Y: 3},
		},
	}

	lastFrame.Snakes = []*pb.Snake{snake}

	gt, err := GameTick(commonGame, lastFrame)
	require.NoError(t, err)
	require.NotNil(t, gt.Snakes[0].Death)
}

func TestUpdateSnakes(t *testing.T) {
	snake := &pb.Snake{
		Body: []*pb.Point{
			{X: 1, Y: 1},
		},
	}
	moves := []*SnakeUpdate{
		{
			Snake: snake,
			Err:   errors.New("some error"),
		},
	}
	updateSnakes(&pb.Game{}, &pb.GameFrame{
		Snakes: []*pb.Snake{snake},
	}, moves)
	require.Equal(t, &pb.Point{X: 1, Y: 0}, snake.Head(), "snake did not move up")

	moves = []*SnakeUpdate{
		{
			Snake: snake,
			Move:  "left",
		},
	}
	updateSnakes(&pb.Game{}, &pb.GameFrame{
		Snakes: []*pb.Snake{snake},
	}, moves)
	require.Equal(t, &pb.Point{X: 0, Y: 0}, snake.Head(), "snake did not move left")
}

func TestCanFollowTail(t *testing.T) {
	url := setupSnakeServer(t, MoveResponse{
		Move: "down",
	}, StartResponse{})
	snake := &pb.Snake{
		Body: []*pb.Point{
			{X: 2, Y: 1},
			{X: 1, Y: 1},
			{X: 1, Y: 2},
			{X: 2, Y: 2},
		},
		URL:    url,
		Health: 100,
	}
	next, err := GameTick(&pb.Game{
		Width:  20,
		Height: 20,
	}, &pb.GameFrame{
		Snakes: []*pb.Snake{snake},
	})
	require.NoError(t, err)
	require.NotNil(t, next)
	require.Nil(t, next.Snakes[0].Death)
}

func TestNextFoodSpawn(t *testing.T) {
	rand.Seed(1) // random order is 65, 85, 29
	snakes := []*pb.Snake{
		{URL: setupSnakeServer(t, MoveResponse{}, StartResponse{})},
		{URL: setupSnakeServer(t, MoveResponse{}, StartResponse{})},
		{URL: setupSnakeServer(t, MoveResponse{}, StartResponse{})},
		{URL: setupSnakeServer(t, MoveResponse{}, StartResponse{})},
	}
	next, err := GameTick(&pb.Game{
		Width:                   20,
		Height:                  20,
		TurnsSinceLastFoodSpawn: 5,
		MaxTurnsToNextFoodSpawn: 5,
	}, &pb.GameFrame{
		Snakes: snakes,
	})
	require.NoError(t, err)
	require.Len(t, next.Food, 0)
}
