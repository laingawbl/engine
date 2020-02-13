package commands

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/laingawbl/engine/rules"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	games int
)

func init() {
	loadTestCmd.Flags().StringVarP(&configFile, "config", "c", "snake-config.json", "specify the location of the snake config file")
	loadTestCmd.Flags().IntVarP(&games, "num-games", "n", 10, "number of games to create and run for the load test")
}

type statusUpdate struct {
	id     string
	status string
}

var loadTestCmd = &cobra.Command{
	Use:   "load-test",
	Short: "run a load test against the engine, using the provided snake config",
	Args: func(c *cobra.Command, args []string) error {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, cr)
		return err
	},
	Run: func(*cobra.Command, []string) {
		start := time.Now()
		ids := []string{}
		log.Info("Creating games")
		for i := 0; i < games; i++ {
			cr := createGame()
			ids = append(ids, cr.ID)
			runGame(cr.ID)
		}

		statuses := map[string]rules.GameStatus{}
		updates := make(chan statusUpdate)
		for _, id := range ids {
			statuses[id] = ""
			go checkStatus(id, updates)
		}

		for s := range updates {
			log.WithFields(log.Fields{
				"id":     s.id,
				"status": s.status,
			}).Info("Game Status")
			statuses[s.id] = rules.GameStatus(s.status)

			done := true
			for _, s := range statuses {
				if s == rules.GameStatusComplete || s == rules.GameStatusError {
					continue
				}
				done = false
			}

			if done {
				log.WithFields(log.Fields{
					"elapsed": time.Since(start),
					"games":   games,
				}).Info("All games complete")
				return
			}
		}
	},
}

var updateFrequency = 300 * time.Millisecond

func checkStatus(id string, updates chan<- statusUpdate) {
	t := time.NewTicker(updateFrequency)
	for range t.C {
		sr := getStatus(id)
		updates <- statusUpdate{id: id, status: sr.Game.Status}
		if sr.Game.Status == string(rules.GameStatusComplete) || sr.Game.Status == string(rules.GameStatusError) {
			t.Stop()
			return
		}
	}
}
