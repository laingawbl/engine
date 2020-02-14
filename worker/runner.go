package worker

import (
	"context"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/battlesnakeio/engine/rules"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	framePushErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "engine",
			Subsystem: "worker",
			Name:      "frames_push_errors",
			Help:      "Number of frames processed by the worker.",
		},
		[]string{},
	)
	framesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "engine",
			Subsystem: "worker",
			Name:      "frames_processed",
			Help:      "Number of frames processed by the worker.",
		},
		[]string{},
	)
)

func init() {
	prometheus.MustRegister(framePushErrors)
	prometheus.MustRegister(framesProcessed)
}

// Runner will run an invidual game to completion. It takes a game id and a
// connection to the controller as arguments.
func Runner(ctx context.Context, client pb.ControllerClient, id string) error {
	resp, err := client.Status(ctx, &pb.StatusRequest{ID: id})
	if err != nil {
		return err
	}
	lastFrame := resp.LastFrame

	for {
		nextFrame, err := rules.GameTick(resp.Game, lastFrame)
		if err != nil {
			// This is a GameFrame error, we can assume that this is a fatal
			// error and no more game processing can take place at this point.
			log.WithError(err).
				WithField("game", id).
				Error("ending game due to fatal error")
			if _, endErr := client.EndGame(ctx, &pb.EndGameRequest{ID: resp.Game.ID}); endErr != nil {
				log.WithError(endErr).
					WithField("game", id).
					Error("failed to end game after fatal error")
			}
			return err
		}

		log.WithField("GameID", id).
			WithField("Turn", nextFrame.Turn).
			Info("adding game frame")
		_, err = client.AddGameFrame(ctx, &pb.AddGameFrameRequest{
			ID:        resp.Game.ID,
			GameFrame: nextFrame,
		})
		if err != nil {
			framePushErrors.With(prometheus.Labels{}).Inc()
			// This is likely a lock error, not to worry here, we can exit.
			return err
		}

		if rules.CheckForGameOver(rules.GameMode(resp.Game.Mode), nextFrame) {
			log.WithField("GameID", id).
				WithField("Turn", nextFrame.Turn).
				Info("ending game")
			rules.NotifyGameEnd(resp.Game, nextFrame)
			_, err := client.EndGame(ctx, &pb.EndGameRequest{ID: resp.Game.ID})
			if err != nil {
				log.WithError(err).WithField("GameID", id).Error("Error while ending game")
			}
			return err
		}

		framesProcessed.With(prometheus.Labels{}).Inc()

		lastFrame = nextFrame
	}
}
