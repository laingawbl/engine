package rules

import (
	"bytes"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/laingawbl/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestNotifyGameEnd(t *testing.T) {
	body := readCloser{Buffer: &bytes.Buffer{}}
	body.WriteString("null")
	urlReceived := ""
	createClient = func(time.Duration) httpClient {
		return mockHTTPClient{
			resp: func(url string) *http.Response {
				urlReceived = url
				return &http.Response{
					Body: body,
				}
			},
		}
	}

	snake := &pb.Snake{
		URL: "http://not.a.snake.com",
	}

	NotifyGameEnd(&pb.Game{}, &pb.GameFrame{
		Snakes: []*pb.Snake{snake},
	})

	require.Equal(t, urlReceived, "http://not.a.snake.com/end")
}

func TestNotifyGameEndSilentlyHandlesError(t *testing.T) {
	createClient = func(time.Duration) httpClient {
		return mockHTTPClient{
			err: errors.New("fail"),
		}
	}

	snake := &pb.Snake{
		URL: "http://not.a.snake.com",
	}

	NotifyGameEnd(&pb.Game{}, &pb.GameFrame{
		Snakes: []*pb.Snake{snake},
	})

	defer func() {
		if r := recover(); r != nil {
			require.Fail(t, "Should not panic on /end error")
		}
	}()
}
