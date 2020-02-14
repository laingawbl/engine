package rules

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	log "github.com/sirupsen/logrus"
)

// NotifyGameEnd sends the /end requests to all the snakes.
func NotifyGameEnd(game *pb.Game, frame *pb.GameFrame) {
	netClient := createClient(200 * time.Millisecond)

	for _, s := range frame.Snakes {
		req := buildSnakeRequest(game, frame, s.ID)
		data, err := json.Marshal(req)

		if err != nil {
			log.WithError(err).WithField("snakeID", s.ID).
				Error("error while marshaling snake request")
			return
		}

		buf := bytes.NewBuffer(data)
		r, err := netClient.Post(getURL(s.URL, "end"), "application/json", buf)
		if err != nil {
			log.WithError(err).WithField("snakeID", s.ID).Error("error POSTing to /end")
		}
		if r != nil {
			if err = r.Body.Close(); err != nil {
				log.WithError(err).Warn("failed to close response body")
			}
		}
	}
}
