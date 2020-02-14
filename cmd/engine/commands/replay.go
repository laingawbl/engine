package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/gorilla/websocket"
	termbox "github.com/nsf/termbox-go"
	"github.com/spf13/cobra"
)

func init() {
	replayCmd.Flags().StringVarP(&gameID, "game-id", "g", "", "the game id of the game to get the status of")
}

var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "replays an existing game on the battlesnake engine",
	Args: func(c *cobra.Command, args []string) error {
		if len(gameID) == 0 {
			return errors.New("game id is required")
		}
		return nil
	},
	Run: func(*cobra.Command, []string) {
		replayGame()
	},
}

func moveFrameForwards(frameIndex int, frames *frameHolder) (int, *pb.GameFrame, bool) {
	frameIndex++
	if frameIndex >= frames.count() {
		return frameIndex, nil, true
	}
	return frameIndex, frames.get(frameIndex), false
}

func moveFrameBackwards(frameIndex int, frames *frameHolder) (int, *pb.GameFrame) {
	frameIndex--
	if frameIndex <= 0 {
		frameIndex = 0
	}
	return frameIndex, frames.get(frameIndex)
}

func loadGame() (*pb.Game, *frameHolder, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("%s/games/%s", apiAddr, gameID))
	if err != nil {
		fmt.Println("error while getting status", err)
		return nil, nil, err
	}
	s := &pb.StatusResponse{}
	err = json.NewDecoder(resp.Body).Decode(s)
	if err != nil {
		fmt.Println("error while getting status", err)
		return nil, nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		fmt.Println("error while closing body", err)
	}

	frames := &frameHolder{}

	u := url.URL{Scheme: "ws", Host: strings.Replace(apiAddr, "http://", "", 1), Path: fmt.Sprintf("/socket/%s", gameID)}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	go func() {
		defer func() {
			err := c.Close()
			if err != nil {
				log.Fatalf("failure to close websocket connection: %v", err)
			}
		}()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				if !strings.Contains(err.Error(), "close 1000 (normal)") {
					log.Println("read:", err)
				}
				return
			}

			switch mt {
			case websocket.TextMessage:
				frame := &pb.GameFrame{}
				err = json.Unmarshal(message, frame)
				if err != nil {
					log.Println("unmarshal frame:", err)
					return
				}

				frames.append(frame)
			case websocket.CloseMessage:
				return
			default:
				log.Println("unhandled message type:", mt)
			}

		}
	}()

	return s.Game, frames, nil
}

func replayGame() {
	game, frames, err := loadGame()
	if err != nil {
		panic(err)
	}

	if err = termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	eventQueue := setupEventQueue()
	currentFrame := getInitialFrame(frames)

	cycle := time.NewTicker(200 * time.Millisecond)
	frameIndex := 0
	paused := false
	done := false

	for {
		if done {
			break
		}

		select {
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey {
				switch ev.Key {
				case termbox.KeyEsc:
					done = true
				case termbox.KeySpace:
					paused = !paused
					if paused {
						cycle.Stop()
					} else {
						cycle = time.NewTicker(200 * time.Millisecond)
					}

				case termbox.KeyArrowLeft:
					paused = true
					frameIndex, currentFrame = moveFrameBackwards(frameIndex, frames)
					if err = render(game, currentFrame); err != nil {
						panic(err)
					}
				case termbox.KeyArrowRight:
					paused = true
					frameIndex, currentFrame, done = moveFrameForwards(frameIndex, frames)
					if err = render(game, currentFrame); err != nil {
						panic(err)
					}
				}

			}
		case <-cycle.C:
			if paused {
				continue
			}
			if err = render(game, currentFrame); err != nil {
				panic(err)
			}
			frameIndex, currentFrame, done = moveFrameForwards(frameIndex, frames)

		}
	}

	if frameIndex >= frames.count() {
		tbprint(0, 0, defaultColor, defaultColor, "Press any key to exit...")
		err = termbox.Flush()
		if err != nil {
			log.Fatalf("Error while flushing termbox: %v", err)
		}
		termbox.PollEvent()
	}
}

func setupEventQueue() <-chan termbox.Event {
	eventQueue := make(chan termbox.Event)
	go func(ev chan<- termbox.Event) {
		for {
			ev <- termbox.PollEvent()
		}
	}(eventQueue)
	return eventQueue
}

func getInitialFrame(frames *frameHolder) *pb.GameFrame {
	var currentFrame *pb.GameFrame
	select {
	case currentFrame = <-frames.initialFrame():
		break
	case <-time.After(100 * time.Millisecond):
		log.Fatal("unable to find initial frame for game")
	}
	return currentFrame
}
