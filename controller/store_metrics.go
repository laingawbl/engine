package controller

import (
	"context"

	"github.com/laingawbl/engine/controller/pb"
	"github.com/laingawbl/engine/rules"
	"github.com/prometheus/client_golang/prometheus"
)

// InstrumentStore wraps all store methods to instrument the underlying calls.
func InstrumentStore(s Store) Store { return &metrics{s} }

var (
	storeCalls = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "engine",
			Subsystem: "worker",
			Name:      "calls",
			Help:      "Calls processed by the store.",
		},
		[]string{"method"},
	)
)

func instrument(method string) func() {
	t := prometheus.NewTimer(storeCalls.WithLabelValues(method))
	return t.ObserveDuration
}

func init() {
	prometheus.MustRegister(storeCalls)
}

type metrics struct{ s Store }

func (m *metrics) Lock(ctx context.Context, key, token string) (string, error) {
	defer instrument("Lock")()
	return m.s.Lock(ctx, key, token)
}

func (m *metrics) Unlock(ctx context.Context, key, token string) error {
	defer instrument("Unlock")()
	return m.s.Unlock(ctx, key, token)
}

func (m *metrics) PopGameID(c context.Context) (string, error) {
	defer instrument("PopGameID")()
	return m.s.PopGameID(c)
}

func (m *metrics) SetGameStatus(c context.Context, id string, status rules.GameStatus) error {
	defer instrument("SetGameStatus")()
	return m.s.SetGameStatus(c, id, status)
}

func (m *metrics) CreateGame(c context.Context, g *pb.Game, frames []*pb.GameFrame) error {
	defer instrument("CreateGame")()
	return m.s.CreateGame(c, g, frames)
}

func (m *metrics) PushGameFrame(c context.Context, id string, t *pb.GameFrame) error {
	defer instrument("PushGameFrame")()
	return m.s.PushGameFrame(c, id, t)
}

func (m *metrics) ListGameFrames(c context.Context, id string, limit, offset int) ([]*pb.GameFrame, error) {
	defer instrument("ListGameFrames")()
	return m.s.ListGameFrames(c, id, limit, offset)
}

func (m *metrics) GetGame(c context.Context, id string) (*pb.Game, error) {
	defer instrument("GetGame")()
	return m.s.GetGame(c, id)
}

func (m *metrics) GameQueueLength(ctx context.Context) (int, int, error) {
	// don't need to instrument this method, since it's used for instrumenting
	return m.s.GameQueueLength(ctx)
}
