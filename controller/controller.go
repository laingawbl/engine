// Package controller provides an API available to workers to write games. It
// also provides the internal API for starting games and watching.
package controller

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/laingawbl/engine/config"
	"github.com/laingawbl/engine/controller/pb"
	"github.com/laingawbl/engine/rules"
	"github.com/laingawbl/engine/version"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	promgrpc "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrLimited = status.Error(codes.ResourceExhausted, "controller: rate limited")

func init() {
	// Enable histogram metrics, they are good.
	promgrpc.EnableHandlingTimeHistogram()
}

var (
	runningGames = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "engine",
			Subsystem: "controller",
			Name:      "games_running",
			Help:      "number of games in a running state",
		})

	gamesWaitingToStart = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "engine",
			Subsystem: "controller",
			Name:      "games_waiting",
			Help:      "number of games that are running, but have no frames yet",
		})
)

func init() {
	prometheus.MustRegister(runningGames, gamesWaitingToStart)
}

// MaxTicks is the maximum amount of ticks that can be returned.
const MaxTicks = 100

// New will initialize a new Server.
func New(store Store) *Server {

	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			running, waiting, err := store.GameQueueLength(context.Background())
			if err != nil {
				log.WithError(err).Error("unable to get game queue length")
				continue
			}
			runningGames.Set(float64(running))
			gamesWaitingToStart.Set(float64(waiting))
		}
	}()
	return &Server{
		Store: store,
		// This effectively sets the limit for games that can be run every
		// second. With the current value we can run 600 games every minute.
		limiter: rate.NewLimiter(config.PopRate, config.PopBurstRate),
		started: make(chan struct{}),
	}
}

// Server is a grpc server for pb.ControllerServer.
type Server struct {
	Store Store

	limiter *rate.Limiter
	started chan struct{}
	port    int
}

// ValidateSnake takes a snake URL and sends requests to validate a snakes validity.
func (s *Server) ValidateSnake(ctx context.Context, req *pb.ValidateSnakeRequest) (*pb.ValidateSnakeResponse, error) {
	url := req.URL

	if url == "" {
		return nil, errors.New("url not found in request")
	}
	gameID := strconv.FormatInt(time.Now().UnixNano(), 10)
	validateSnakeResponse := &pb.ValidateSnakeResponse{
		StartStatus: rules.ValidateStart(gameID, url, rules.SlowSnakeMS),
		MoveStatus:  rules.ValidateMove(gameID, url, rules.SlowSnakeMS),
		EndStatus:   rules.ValidateEnd(gameID, url, rules.SlowSnakeMS),
		PingStatus:  rules.ValidatePing("nogame", url, rules.SlowSnakeMS),
	}
	return validateSnakeResponse, nil
}

// Pop should pop a game that is unlocked and unfinished from the queue, lock
// the game and return it to the worker to begin processing. This call will
// be polled by the workers.
func (s *Server) Pop(ctx context.Context, _ *pb.PopRequest) (*pb.PopResponse, error) {
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, ErrLimited
	}

	id, err := s.Store.PopGameID(ctx)
	if err != nil {
		return nil, err
	}

	token, err := s.Store.Lock(ctx, id, "")
	if err != nil {
		return nil, err
	}

	return &pb.PopResponse{ID: id, Token: token}, nil
}

// Status retrieves the game state including the last processed game frame.
func (s *Server) Status(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	game, err := s.Store.GetGame(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	var lastFrame *pb.GameFrame
	frames, err := s.Store.ListGameFrames(ctx, req.ID, 1, -1)
	if err != nil {
		return nil, err
	}
	if len(frames) > 0 {
		lastFrame = frames[0]
	}
	return &pb.StatusResponse{Game: game, LastFrame: lastFrame}, nil
}

// Start starts the game running, and will make it ready to be picked up by a
// worker.
func (s *Server) Start(ctx context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	g, err := s.Store.GetGame(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if rules.GameStatus(g.Status) == rules.GameStatusComplete || rules.GameStatus(g.Status) == rules.GameStatusError {
		return nil, status.Error(codes.InvalidArgument, "cannot start a game that is in complete or error state")
	}
	err = s.Store.SetGameStatus(ctx, req.ID, rules.GameStatusRunning)
	if err != nil {
		return nil, err
	}
	return &pb.StartResponse{}, nil
}

// Create creates a new game, but doesn't start running frames.
func (s *Server) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	game, frames, err := rules.CreateInitialGame(req)
	if err != nil {
		return nil, err
	}
	err = s.Store.CreateGame(ctx, game, frames)
	if err != nil {
		return nil, err
	}
	return &pb.CreateResponse{
		ID: game.ID,
	}, nil
}

// AddGameFrame adds a new game frame to the game. A lock must be held for this
// call to succeed.
func (s *Server) AddGameFrame(ctx context.Context, req *pb.AddGameFrameRequest) (*pb.AddGameFrameResponse, error) {
	token := pb.ContextGetLockToken(ctx)

	if req.GameFrame == nil {
		return nil, status.Error(codes.InvalidArgument, "controller: game frame must not be nil")
	}

	// Lock the game again, if this fails, the lock is not valid.
	_, err := s.Store.Lock(ctx, req.ID, token)
	if err != nil {
		return nil, err
	}

	err = s.Store.PushGameFrame(ctx, req.ID, req.GameFrame)
	if err != nil {
		return nil, err
	}
	game, err := s.Store.GetGame(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return &pb.AddGameFrameResponse{
		Game: game,
	}, nil
}

// ListGameFrames will list all game frames given a limit and offset.
func (s *Server) ListGameFrames(ctx context.Context, req *pb.ListGameFramesRequest) (*pb.ListGameFramesResponse, error) {
	if req.Limit == 0 || req.Limit >= MaxTicks {
		req.Limit = MaxTicks
	}
	frames, err := s.Store.ListGameFrames(ctx, req.ID, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}
	return &pb.ListGameFramesResponse{
		Frames: frames,
		Count:  int32(len(frames)),
	}, nil
}

// EndGame sets the game status to complete. A lock must be held for this call
// to succeed.
func (s *Server) EndGame(ctx context.Context, req *pb.EndGameRequest) (*pb.EndGameResponse, error) {
	token := pb.ContextGetLockToken(ctx)

	// Lock the game again, if this fails, the lock is not valid. We only need
	// the lock for the next part where we set the game status.
	newToken, err := s.Store.Lock(ctx, req.ID, token)
	if err != nil {
		return nil, err
	}
	token = newToken

	err = s.Store.SetGameStatus(ctx, req.ID, rules.GameStatusComplete)
	if err != nil {
		return nil, err
	}

	err = s.Store.Unlock(ctx, req.ID, token)
	if err != nil {
		return nil, err
	}

	return &pb.EndGameResponse{}, nil
}

// Ping returns the health and current version of the server.
func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Version: version.Version}, nil
}

// Serve will instantiate a grpc server.
func (s *Server) Serve(listen string) error {
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	s.port = lis.Addr().(*net.TCPAddr).Port
	srv := grpc.NewServer(grpc.UnaryInterceptor(
		grpcmiddleware.ChainUnaryServer(
			loggingInterceptor,
			promgrpc.UnaryServerInterceptor,
		),
	))
	pb.RegisterControllerServer(srv, s)
	close(s.started)
	return srv.Serve(lis)
}

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	if info.FullMethod != "/pb.Controller/Pop" {
		log.WithFields(log.Fields{
			"duration": time.Since(start),
			"method":   info.FullMethod,
		}).Info("controller rpc")
	}
	return resp, err
}

// DialAddress will return a localhost address to reach the server. This is
// useful if the server will select it's own port.
func (s *Server) DialAddress() string {
	s.Wait()
	return fmt.Sprintf("127.0.0.1:%d", s.port)
}

// Wait will wait until the server has started.
func (s *Server) Wait() { <-s.started }
