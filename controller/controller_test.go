package controller

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/battlesnakeio/engine/rules"
	"github.com/battlesnakeio/engine/version"
	"github.com/stretchr/testify/require"
)

var client pb.ControllerClient
var store Store

func init() {
	store = InMemStore()
	ctrl := New(store)
	go func() {
		if err := ctrl.Serve(":0"); err != nil {
			panic(err)
		}
	}()
	var err error
	client, err = pb.Dial(ctrl.DialAddress())
	if err != nil {
		panic(err)
	}
}

func TestController_GameSnakeTimeouts(t *testing.T) {
	ctx := context.Background()

	t.Run("GetGame_SnakeTimeoutSet", func(t *testing.T) {
		resp, _ := client.Create(ctx, &pb.CreateRequest{SnakeTimeout: 250})
		game, _ := client.Status(ctx, &pb.StatusRequest{ID: resp.ID})
		require.Equal(t, int32(250), game.Game.SnakeTimeout)
	})

	t.Run("GetGame_SnakeTimeoutSet0", func(t *testing.T) {
		resp, _ := client.Create(ctx, &pb.CreateRequest{SnakeTimeout: 0})
		game, _ := client.Status(ctx, &pb.StatusRequest{ID: resp.ID})
		require.Equal(t, int32(500), game.Game.SnakeTimeout)
	})

	t.Run("GetGame_SnakeTimeoutSet6000", func(t *testing.T) {
		resp, _ := client.Create(ctx, &pb.CreateRequest{SnakeTimeout: 6000})
		game, _ := client.Status(ctx, &pb.StatusRequest{ID: resp.ID})
		require.Equal(t, int32(500), game.Game.SnakeTimeout)
	})
}
func TestController_GameCRUD(t *testing.T) {
	ctx := context.Background()

	var gameID string
	var token string

	t.Run("StartGame_NoGame", func(t *testing.T) {
		_, err := client.Start(ctx, &pb.StartRequest{ID: gameID})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = NotFound desc = controller: game not found",
			err.Error(),
		)
	})

	t.Run("Create", func(t *testing.T) {
		resp, err := client.Create(ctx, &pb.CreateRequest{})
		require.Nil(t, err)
		gameID = resp.ID
	})

	t.Run("StartGame", func(t *testing.T) {
		_, err := client.Start(ctx, &pb.StartRequest{ID: gameID})
		require.Nil(t, err)
	})

	t.Run("PopGame", func(t *testing.T) {
		resp, err := client.Pop(ctx, &pb.PopRequest{})
		require.Nil(t, err)
		require.Equal(t, gameID, resp.ID)
		token = resp.Token
	})

	t.Run("PopGame_NoGames", func(t *testing.T) {
		_, err := client.Pop(ctx, &pb.PopRequest{})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = NotFound desc = controller: game not found",
			err.Error(),
		)
	})

	t.Run("GetGame", func(t *testing.T) {
		game, err := client.Status(ctx, &pb.StatusRequest{ID: gameID})
		require.Nil(t, err)
		require.Equal(t, "running", game.Game.Status)
		require.Equal(t, int32(500), game.Game.SnakeTimeout)
		require.Equal(t, "multi-player", game.Game.Mode)
		require.NotNil(t, game.LastFrame)
		require.Equal(t, int32(0), game.LastFrame.Turn)
	})

	t.Run("GetGame_NoGames", func(t *testing.T) {
		_, err := client.Status(ctx, &pb.StatusRequest{
			ID: "non-existant",
		})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = NotFound desc = controller: game not found",
			err.Error(),
		)
	})

	t.Run("EndGame_NoToken", func(t *testing.T) {
		_, err := client.EndGame(ctx, &pb.EndGameRequest{ID: gameID})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = ResourceExhausted desc = controller: game is locked",
			err.Error(),
		)
	})

	t.Run("EndGame", func(t *testing.T) {
		_, err := client.EndGame(
			pb.ContextWithLockToken(ctx, token), &pb.EndGameRequest{ID: gameID})
		require.Nil(t, err)
		g, err := store.GetGame(ctx, gameID)
		require.NoError(t, err)
		require.Equal(t, rules.GameStatusComplete, rules.GameStatus(g.Status))
	})

	t.Run("StartGameOnCompletedGame", func(t *testing.T) {
		_, err := client.Start(ctx, &pb.StartRequest{ID: gameID})
		require.Error(t, err)
	})
}

func TestController_Frames(t *testing.T) {
	ctx := context.Background()

	var gameID string
	var token string

	t.Run("Create", func(t *testing.T) {
		resp, err := client.Create(ctx, &pb.CreateRequest{})
		require.Nil(t, err)
		gameID = resp.ID
	})

	t.Run("StartGame", func(t *testing.T) {
		_, err := client.Start(ctx, &pb.StartRequest{ID: gameID})
		require.Nil(t, err)
	})

	t.Run("PopGame", func(t *testing.T) {
		resp, err := client.Pop(ctx, &pb.PopRequest{})
		require.Nil(t, err)
		require.Equal(t, gameID, resp.ID)
		token = resp.Token
	})

	t.Run("AddGameFrame_NoGame", func(t *testing.T) {
		_, err := client.AddGameFrame(ctx, &pb.AddGameFrameRequest{
			ID:        "foo",
			GameFrame: &pb.GameFrame{},
		})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = NotFound desc = controller: game not found",
			err.Error(),
		)
	})

	t.Run("AddGameFrame_NoLock", func(t *testing.T) {
		_, err := client.AddGameFrame(ctx, &pb.AddGameFrameRequest{
			ID:        gameID,
			GameFrame: &pb.GameFrame{},
		})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = ResourceExhausted desc = controller: game is locked",
			err.Error(),
		)
	})

	t.Run("AddGameFrame_Nil", func(t *testing.T) {
		_, err := client.AddGameFrame(
			pb.ContextWithLockToken(ctx, token),
			&pb.AddGameFrameRequest{
				ID:        gameID,
				GameFrame: nil,
			})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = InvalidArgument desc = controller: game frame must not be nil",
			err.Error(),
		)
	})

	t.Run("AddGameFrames", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			resp, err := client.AddGameFrame(
				pb.ContextWithLockToken(ctx, token), &pb.AddGameFrameRequest{
					ID: gameID,
					GameFrame: &pb.GameFrame{
						Turn: int32(i + 1),
					},
				})
			require.Nil(t, err)
			require.Equal(t, gameID, resp.Game.ID)
		}
	})

	t.Run("ListGameFrames_NoGame", func(t *testing.T) {
		_, err := client.ListGameFrames(ctx, &pb.ListGameFramesRequest{ID: "foo"})
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = NotFound desc = controller: game not found",
			err.Error(),
		)
	})

	t.Run("ListGameFrames_BadOffset", func(t *testing.T) {
		resp, err := client.ListGameFrames(ctx, &pb.ListGameFramesRequest{
			ID: gameID, Offset: 1000000})
		require.Nil(t, err)
		require.Empty(t, resp.Frames)
	})

	t.Run("ListGameFrames", func(t *testing.T) {
		resp, err := client.ListGameFrames(ctx, &pb.ListGameFramesRequest{ID: gameID})
		require.Nil(t, err)
		require.Equal(t, MaxTicks, len(resp.Frames))
		require.Equal(t, MaxTicks, int(resp.Count))
	})

	t.Run("ListGameFramesAtMaxTicks", func(t *testing.T) {
		resp, err := client.ListGameFrames(ctx, &pb.ListGameFramesRequest{
			ID:    gameID,
			Limit: 1000,
		})
		require.Nil(t, err)
		require.Equal(t, MaxTicks, len(resp.Frames))
		require.Equal(t, MaxTicks, int(resp.Count))
	})
}

func TestController_PopConcurrent(t *testing.T) {
	ctx := context.Background()
	ctrl := client

	g, _ := ctrl.Create(ctx, &pb.CreateRequest{})
	_, err := ctrl.Start(ctx, &pb.StartRequest{ID: g.ID})
	require.Nil(t, err)

	var ok uint32 // How many got the lock.
	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			res, errp := ctrl.Pop(ctx, &pb.PopRequest{})
			if errp == nil {
				fmt.Printf("res: %v -- %+v\n", time.Now(), res)
				atomic.AddUint32(&ok, 1)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	require.Equal(t, uint32(1), ok)
}

func TestController_Ping(t *testing.T) {
	ctx := context.Background()

	res, err := client.Ping(ctx, &pb.PingRequest{})
	require.Nil(t, err)

	require.Equal(t, version.Version, res.Version)
}

func TestController_ValidateSnakeNoURL(t *testing.T) {
	_, err := client.ValidateSnake(context.Background(), &pb.ValidateSnakeRequest{})
	require.Error(t, err, "Expected error with no URL")
}

func TestController_ValidateSnakeInvalidURL(t *testing.T) {
	res, err := client.ValidateSnake(context.Background(), &pb.ValidateSnakeRequest{
		URL: "aoeu",
	})
	require.Nil(t, err)
	require.Equal(t, "Snake URL not valid", res.StartStatus.Message)
}

func TestController_ValidateSnakeValidUrlNoServer(t *testing.T) {
	res, err := client.ValidateSnake(context.Background(), &pb.ValidateSnakeRequest{
		URL: "http://shouldneverresolveinamillionyearsaoeu.com",
	})
	require.Nil(t, err)
	require.True(t, strings.Index(res.StartStatus.Errors[0], "Post http://shouldneverresolveinamillionyearsaoeu.com/start: dial tcp: lookup shouldneverresolveinamillionyearsaoeu.com") == 0, "Found unexpected string: "+res.StartStatus.Errors[0])
}
