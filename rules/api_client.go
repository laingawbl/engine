package rules

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	snakeRequestsHistogramMetric = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "engine",
			Subsystem: "worker",
			Name:      "snake_requests_duration",
			Help:      "Calls to outbound snakes.",
		},
		[]string{"method", "code", "official"},
	)
)

func init() { prometheus.MustRegister(snakeRequestsHistogramMetric) }

// Official snake url is set for tracking purposes. If requests are failing to
// this snake then things are going wrong!
var officialSnakeURL = os.Getenv("OFFICIAL_SNAKE_URL")

type snakeResponse struct {
	snake   *pb.Snake
	data    []byte
	err     error
	latency time.Duration
}

type multiSnakeRequest struct {
	url     string
	timeout time.Duration
	game    *pb.Game
	frame   *pb.GameFrame
}

type snakePostOptions struct {
	url     string
	snake   *pb.Snake
	timeout time.Duration
}

type snakePostRequest struct {
	options snakePostOptions
	data    []byte
}

func gatherAllSnakeResponses(multiReq multiSnakeRequest) []snakeResponse {
	return gatherSnakeResponses(multiReq, multiReq.frame.Snakes)
}

func gatherAliveSnakeResponses(multiReq multiSnakeRequest) []snakeResponse {
	return gatherSnakeResponses(multiReq, multiReq.frame.AliveSnakes())
}

func gatherSnakeResponses(multiReq multiSnakeRequest, snakes []*pb.Snake) []snakeResponse {
	respChan := make(chan snakeResponse, len(multiReq.frame.Snakes))
	wg := sync.WaitGroup{}

	for _, snake := range snakes {
		if !isValidURL(snake.URL) {
			respChan <- snakeResponse{
				snake: snake,
				err:   errors.New("invalid snake URL: " + snake.URL),
			}
			continue
		}

		wg.Add(1)
		go func(s *pb.Snake, mr multiSnakeRequest) {
			options := snakePostOptions{
				url:     mr.url,
				snake:   s,
				timeout: mr.timeout,
			}
			getSnakeResponse(options, mr.game, mr.frame, respChan)
			wg.Done()
		}(snake, multiReq)
	}

	wg.Wait()
	close(respChan)

	ret := []snakeResponse{}
	for response := range respChan {
		ret = append(ret, response)
	}
	return ret
}

func postToSnakeServer(req snakePostRequest, resp chan<- snakeResponse) {
	instrument := func(latency time.Duration, statusCode int) {
		instrumentSnakeCall(req.options.url, req.options.snake.URL == officialSnakeURL, statusCode, latency)
	}

	buf := bytes.NewBuffer(req.data)
	netClient := createClient(req.options.timeout)
	postURL := getURL(req.options.snake.URL, req.options.url)

	start := time.Now()
	postResponse, err := netClient.Post(postURL, "application/json", buf)
	latency := time.Since(start)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"url": postURL,
			"id":  req.options.snake.ID,
		}).Error("error POSTing to snake")
		resp <- snakeResponse{
			snake: req.options.snake,
			err:   err,
		}
		instrument(0, 0)
		return
	}

	defer func() {
		if bErr := postResponse.Body.Close(); bErr != nil {
			log.WithError(bErr).Warn("failed to close response body")
		}
	}()
	instrument(latency, postResponse.StatusCode)

	// Limited read to 1mb of data.
	responseData, err := ioutil.ReadAll(io.LimitReader(postResponse.Body, 1000000))

	resp <- snakeResponse{
		snake:   req.options.snake,
		data:    responseData,
		err:     err,
		latency: latency,
	}
}

func getSnakeResponse(options snakePostOptions, game *pb.Game, frame *pb.GameFrame, resp chan<- snakeResponse) {
	req := buildSnakeRequest(game, frame, options.snake.ID)
	data, err := json.Marshal(req)

	if err != nil {
		log.WithError(err).WithField("snakeID", options.snake.ID).
			Error("error while marshaling snake request")
		resp <- snakeResponse{
			snake: options.snake,
			err:   err,
		}
		return
	}

	postToSnakeServer(snakePostRequest{
		options: options,
		data:    data,
	}, resp)
}

func instrumentSnakeCall(method string, official bool, statusCode int, latency time.Duration) {
	var status string
	if statusCode >= 200 {
		status = "2xx"
	} else if statusCode >= 300 {
		status = "3xx"
	} else if statusCode >= 400 {
		status = "4xx"
	} else if statusCode >= 500 {
		status = "5xx"
	} else {
		status = "err"
	}
	snakeRequestsHistogramMetric.WithLabelValues(method, status, fmt.Sprint(official)).Observe(
		latency.Seconds(),
	)
}
