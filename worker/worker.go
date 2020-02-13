// Package worker provides the actual running of games. It is the core of the
// engine and game logic. It interacts primarily with the controller API writing
// and reading game state.
package worker

import (
	"context"
	"time"

	"github.com/laingawbl/engine/controller/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Worker is the worker interface. It wraps a main Perform functions which is
// where all of the game logic should live.
type Worker struct {
	ControllerClient pb.ControllerClient
	PollInterval     time.Duration
	RunGame          func(context.Context, pb.ControllerClient, string) error
}

// Run will run the worker in a loop.
func (w *Worker) Run(ctx context.Context, workerID int) {
	for {
		// We are now holding the lock.
		if err := w.run(ctx, workerID); err != nil {
			s, ok := status.FromError(err)
			if !ok || s.Code() != codes.NotFound {
				log.WithError(err).WithField("worker", workerID).Warn("run failed")
			}

			select {
			case <-time.After(w.PollInterval):
			case <-ctx.Done():
				return
			}
		}
	}
}

func (w *Worker) run(ctx context.Context, workerID int) error {
	// Pop an item of work.
	pop, err := w.ControllerClient.Pop(ctx, &pb.PopRequest{})
	if err != nil {
		return err
	}

	log.WithField("worker", workerID).
		WithField("game", pop.ID).
		Info("acquired lock")

	// Get a context with the lock token.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx = pb.ContextWithLockToken(ctx, pop.Token)

	// Perform the actual work, this should respect context and Done() rules.
	// Perform should be able to write to storage using the context and have
	// a valid lock for the key.
	return w.RunGame(ctx, w.ControllerClient, pop.ID)
}
