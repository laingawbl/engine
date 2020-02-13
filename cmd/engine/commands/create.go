package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/laingawbl/engine/controller/pb"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "creates a new game on the battlesnake engine",
	Args: func(c *cobra.Command, args []string) error {
		data, err := ioutil.ReadFile(configFile) // nolint: gosec
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, cr)
		return err
	},
	Run: func(*cobra.Command, []string) {
		cr := createGame()
		if cr != nil {
			fmt.Printf(`{"ID": "%s"}`, cr.ID)
		}
	},
}

func createGame() *pb.CreateResponse {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	data, err := json.Marshal(cr)
	if err != nil {
		fmt.Println("unable to marshal request", err)
		return nil
	}
	buf := bytes.NewBuffer(data)
	resp, err := client.Post(fmt.Sprintf("%s/games", apiAddr), "application/json", buf)
	if err != nil {
		fmt.Println("error while posting to create endpoint", err)
		return nil
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("unable to read response body", err)
		return nil
	}

	cr := &pb.CreateResponse{}
	err = json.Unmarshal(data, &cr)
	if err != nil {
		fmt.Println("Unable to unmarshal create response")
	}
	return cr
}

var (
	configFile string
	cr         = &pb.CreateRequest{}
)

func init() {
	createCmd.Flags().StringVarP(&configFile, "config", "c", "snake-config.json", "specify the location of the snake config file")
}
