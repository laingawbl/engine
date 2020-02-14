package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/battlesnakeio/engine/controller/pb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "gets the status of a game from the battlesnake engine",
	Args: func(c *cobra.Command, args []string) error {
		if len(gameID) == 0 {
			return errors.New("game id is required")
		}
		return nil
	},
	Run: func(*cobra.Command, []string) {
		spew.Dump(getStatus(gameID))
	},
}

var (
	gameID string
)

func init() {
	statusCmd.Flags().StringVarP(&gameID, "game-id", "g", "", "the game id of the game to get the status of")
}

func getStatus(id string) *pb.StatusResponse {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("%s/games/%s", apiAddr, id))
	if err != nil {
		fmt.Println("error while posting to status endpoint", err)
		return nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("unable to read response body", err)
		return nil
	}

	sr := &pb.StatusResponse{}
	err = json.Unmarshal(data, sr)
	if err != nil {
		log.WithFields(log.Fields{
			"resp": string(data),
			"id":   id,
		}).Infof("unable to unmarshal status response: %s", string(data))
		return nil
	}

	return sr
}
