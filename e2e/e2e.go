package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/battlesnakeio/engine/controller/pb"
)

type client struct {
	apiURL string
	client *http.Client
}

func (c *client) beginGame(cr *pb.CreateRequest) (string, error) {
	var gameID string

	{
		data, err := json.Marshal(cr)
		if err != nil {
			return "", err
		}
		buf := bytes.NewBuffer(data)
		resp, err := c.client.Post(fmt.Sprintf("%s/games", c.apiURL), "application/json", buf)
		if err != nil {
			return "", err
		}
		res := &pb.CreateResponse{}
		err = json.NewDecoder(resp.Body).Decode(res)
		if cErr := resp.Body.Close(); err != nil {
			return "", cErr
		}
		if err != nil {
			return "", err
		}
		gameID = res.ID
	}

	{
		resp, err := c.client.Post(fmt.Sprintf("%s/games/%s/start", c.apiURL, gameID), "application/json", nil)
		if err != nil {
			return "", err
		}
		err = resp.Body.Close()
		if err != nil {
			return "", err
		}
	}

	return gameID, nil
}

func (c *client) gameStatus(gameID string) (*pb.StatusResponse, *pb.ListGameFramesResponse, error) {
	st := &pb.StatusResponse{}
	frames := &pb.ListGameFramesResponse{}

	{
		resp, err := c.client.Get(fmt.Sprintf("%s/games/%s", c.apiURL, gameID))
		if err != nil {
			return nil, nil, err
		}
		err = json.NewDecoder(resp.Body).Decode(st)
		if err != nil {
			return nil, nil, err
		}
		err = resp.Body.Close()
		if err != nil {
			return nil, nil, err
		}
	}
	{
		resp, err := c.client.Get(fmt.Sprintf("%s/games/%s/frames", c.apiURL, gameID))
		if err != nil {
			return nil, nil, err
		}
		err = json.NewDecoder(resp.Body).Decode(frames)
		if err != nil {
			return nil, nil, err
		}
		err = resp.Body.Close()
		if err != nil {
			return nil, nil, err
		}
	}
	return st, frames, nil
}
