package rules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	log "github.com/sirupsen/logrus"
)

const (
	// SlowSnakeMS represents what is considered a slow response time to the Validate* calls.
	SlowSnakeMS = 1000
)

// ValidateStart validates the start end point on a snake server
func ValidateStart(gameID string, url string, slowSnakeMS int32) *pb.SnakeResponseStatus {
	response := scoreResponse(gameID, url, "start", slowSnakeMS)
	return response
}

// ValidateMove validates the move end point on a snake server
func ValidateMove(gameID string, url string, slowSnakeMS int32) *pb.SnakeResponseStatus {
	response := scoreResponse(gameID, url, "move", slowSnakeMS)
	return response
}

// ValidateEnd validates the end end point on a snake server
func ValidateEnd(gameID string, url string, slowSnakeMS int32) *pb.SnakeResponseStatus {
	response := scoreResponse(gameID, url, "end", slowSnakeMS)
	return response
}

// ValidatePing validates the ping endpoint on a snake server
func ValidatePing(gameID string, url string, slowSnakeMS int32) *pb.SnakeResponseStatus {
	response := scoreResponse(gameID, url, "ping", slowSnakeMS)
	return response
}

func scoreResponse(gameID string, url string, endpoint string, slowSnakeMS int32) *pb.SnakeResponseStatus {
	game, frame := createGameFrame(gameID, url)
	response := &pb.SnakeResponseStatus{
		Score: &pb.Score{},
	}
	if !isValidURL(url) {
		response.Score.ChecksFailed++
		response.Message = "Snake URL not valid"
		response.Errors = []string{"invalid url '" + url + "'"}
	} else {
		response.Score.ChecksPassed++
		rawResponse, responseCode, responseTime, responseError := makeSnakeCall(game, frame, url, endpoint)
		response.Message = "Perfect"
		response.Raw = rawResponse
		response.Time = responseTime
		response.StatusCode = int32(responseCode)
		if response.StatusCode > 0 && (response.StatusCode < 200 || response.StatusCode >= 300) {
			response.Score.ChecksFailed++
			response.Message = "Bad return code, expected 200"
			response.Errors = append(response.Errors, fmt.Sprintf("incorrect http response code, got %d, expected 200", response.StatusCode))
		} else {
			response.Score.ChecksPassed++
		}
		if responseError != nil {
			response.Score.ChecksFailed++
			if strings.HasPrefix(responseError.Error(), "invalid") {
				response.Message = "Bad response format - please ensure you return valid JSON"
				response.Errors = append(response.Errors, responseError.Error())
			} else {
				response.Message = "Unknown error"
				response.Errors = append(response.Errors, responseError.Error())
			}
		} else {
			response.Score.ChecksPassed++
		}
		if responseTime < slowSnakeMS {
			response.Score.ChecksPassed++
		} else {
			response.Message = "Slow snake"
			response.Errors = append(response.Errors, fmt.Sprintf("snake took %d ms to respond, try and get it < %d ms.", responseTime, slowSnakeMS))
			response.Score.ChecksFailed++
		}

	}
	return response
}

func makeSnakeCall(game *pb.Game, frame *pb.GameFrame, url string, endpoint string) (string, int, int32, error) {
	netClient := createClient(5000 * time.Millisecond)
	req := buildSnakeRequest(game, frame, "you")

	data, err := json.Marshal(req)
	if err != nil {
		return "", 0, 0, err
	}
	if endpoint == "ping" {
		data = []byte("{}")
	}
	buf := bytes.NewBuffer(data)
	start := time.Now().UnixNano()
	response, err := netClient.Post(getURL(url, endpoint), "application/json", buf)
	if err != nil {
		return "", 0, 0, err
	}
	defer func() {
		err = response.Body.Close()
		if err != nil {
			log.WithError(err).Error("Unable to close websocket stream")
		}
	}()
	statusCode := response.StatusCode
	finish := time.Now().UnixNano()
	time := int32((finish - start) / int64(time.Millisecond))
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", statusCode, 0, err
	}
	raw := string(contents)

	if endpoint != "end" && endpoint != "ping" {
		var raw map[string]interface{}
		err = json.Unmarshal(contents, &raw)
	}

	return raw, statusCode, time, err
}

func createGameFrame(gameID string, url string) (*pb.Game, *pb.GameFrame) {
	game := &pb.Game{
		Height: 10,
		Width:  10,
		ID:     gameID,
		Status: "Running",
	}
	snake := &pb.Snake{
		ID:   "you",
		URL:  url,
		Name: "you",
		Body: []*pb.Point{
			&pb.Point{X: 2, Y: 2},
		},
	}
	frame := &pb.GameFrame{
		Snakes: []*pb.Snake{
			snake,
		},
		Turn: 1,
		Food: []*pb.Point{&pb.Point{
			X: 1,
			Y: 1,
		},
		},
	}
	return game, frame
}
