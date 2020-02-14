package e2e

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func newClient(url string) *client {
	return &client{
		apiURL: url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

var snakeURL string

func init() {
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	tst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		directions := []string{
			"up",
			"down",
			"left",
			"right",
		}
		direction := directions[randGen.Intn(4)]
		_, err := w.Write([]byte(`{"move":"` + direction + `"}`))
		if err != nil {
			fmt.Println("unable to write bytes")
			os.Exit(1)
		}
	}))
	snakeURL = tst.URL
}

var games = map[string]*pb.CreateRequest{
	"Simple": {
		Width:  5,
		Height: 5,
		Food:   5,
		Snakes: []*pb.SnakeOptions{
			{
				Name: "1",
				URL:  snakeURL,
				ID:   "1",
			},
			{
				Name: "2",
				URL:  snakeURL,
				ID:   "2",
			},
		},
	},
	"InvalidURL": {
		Width:  5,
		Height: 5,
		Food:   5,
		Snakes: []*pb.SnakeOptions{
			{
				Name: "1",
				URL:  snakeURL,
				ID:   "1",
			},
			{
				Name: "2",
				URL:  "invalid",
				ID:   "2",
			},
		},
	},
	"LargerBoard": {
		Width:  100,
		Height: 100,
		Food:   50,
		Snakes: []*pb.SnakeOptions{
			{
				Name: "1",
				URL:  snakeURL,
			},
			{
				Name: "2",
				URL:  snakeURL,
			},
		},
	},
}

func TestMain(m *testing.M) {
	enableE2e := flag.Bool("enable-e2e", false, "enable e2e tests")
	flag.Parse()

	if !*enableE2e {
		os.Exit(0)
		return
	}

	proc := exec.Command("engine", "server", "--chaos")
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr

	if err := proc.Start(); err != nil {
		panic(err)
	}

	code := m.Run()

	err := proc.Process.Kill()
	if err != nil {
		fmt.Printf("error while killing process: %v\n", err)
	}
	os.Exit(code)
}

func Test(t *testing.T) {
	const (
		multiplier   = 5
		waitTicks    = 60
		waitInterval = 1 * time.Second
	)

	apiURL := "http://127.0.0.1:3005"
	c := newClient(apiURL)

	for i := 0; i < 10; i++ {
		if _, getErr := http.Get(apiURL); getErr == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	for i := 0; i < multiplier; i++ {
		for name, game := range games {
			t.Run(fmt.Sprintf("%s#%d", name, i), func(t *testing.T) {
				t.Parallel()

				id, err := c.beginGame(game)
				if !assert.Nil(t, err) {
					return
				}

				var st *pb.StatusResponse
				var ticks *pb.ListGameFramesResponse
				for i := 0; i < waitTicks; i++ {
					time.Sleep(waitInterval)
					st, ticks, err = c.gameStatus(id)
					if !assert.Nil(t, err) {
						return
					}

					if st.Game.Status == "complete" {
						t.Logf("game finished id=%s turns=%d ticks=%d", id, st.LastFrame.Turn, len(ticks.Frames))
						if !assert.Equal(t, int(st.LastFrame.Turn)+1, len(ticks.Frames)) {
							spew.Dump(ticks.Frames)
						}
						for i, tk := range ticks.Frames {
							assert.Equal(t, i, int(tk.Turn))
						}
						return
					}
				}

				spew.Dump(st)
				t.Errorf("test failed after: %v", time.Duration(waitTicks)*waitInterval)
			})
		}
	}
}
