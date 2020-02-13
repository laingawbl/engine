package rules

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/laingawbl/engine/controller/pb"
	log "github.com/sirupsen/logrus"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// GameTick runs the game one tick and updates the state
func GameTick(game *pb.Game, lastFrame *pb.GameFrame) (*pb.GameFrame, error) {
	if lastFrame == nil {
		return nil, fmt.Errorf("rules: invalid state, previous frame is nil")
	}
	nextFrame := &pb.GameFrame{
		Turn:   lastFrame.Turn + 1,
		Snakes: lastFrame.Snakes,
		Food:   lastFrame.Food,
	}
	duration := time.Duration(game.SnakeTimeout) * time.Millisecond
	log.WithFields(log.Fields{
		"GameID":  game.ID,
		"Turn":    nextFrame.Turn,
		"Timeout": duration,
	}).Info("GatherSnakeMoves")
	moves := GatherSnakeMoves(duration, game, lastFrame)

	// we have all the snake moves now
	// 1. snake update
	//    a - update snake coords
	//    b - reverse snakes who have moved backward onto their own bodies
	//    c - remove snake tails
	updateSnakes(game, nextFrame, moves)
	// 2. game update
	//    a - turn incr -- done above when the next tick is created
	//    b - reduce health points
	log.WithFields(log.Fields{
		"GameID": game.ID,
		"Turn":   nextFrame.Turn,
	}).Info("reduce snake health")
	for _, s := range nextFrame.AliveSnakes() {
		s.Health = s.Health - 1
	}

	// 3. check for death
	// 	  a - starvation
	//    b - wall collision
	//    c - snake collision
	log.WithFields(log.Fields{
		"GameID": game.ID,
		"Turn":   nextFrame.Turn,
	}).Info("check for death")
	deathUpdates := checkForDeath(game.Width, game.Height, nextFrame)
	for _, du := range deathUpdates {
		if du.Snake.Death == nil {
			du.Snake.Death = du.Death
		}
	}
	return nextFrame, nil
}

func getUnoccupiedPoint(width, height int32, food []*pb.Point, snakes []*pb.Snake) *pb.Point {
	openPoints := getUnoccupiedPoints(width, height, food, snakes)
	return pickRandomPoint(openPoints)
}

func getUnoccupiedPointOdd(width, height int32, food []*pb.Point, snakes []*pb.Snake) *pb.Point {
	openPoints := getUnoccupiedPoints(width, height, food, snakes)
	openPoints = filterPoints(openPoints, true)
	return pickRandomPoint(openPoints)
}

func getUnoccupiedPointEven(width, height int32, food []*pb.Point, snakes []*pb.Snake) *pb.Point {
	openPoints := getUnoccupiedPoints(width, height, food, snakes)
	openPoints = filterPoints(openPoints, false)
	return pickRandomPoint(openPoints)
}

func filterPoints(openPoints []*pb.Point, even bool) []*pb.Point {
	filteredPoints := []*pb.Point{}
	mod := int32(0)
	if !even {
		mod = int32(1)
	}
	for i := int32(0); i < int32(len(openPoints)); i++ {
		if (openPoints[i].X+openPoints[i].Y)%2 != mod {
			filteredPoints = append(filteredPoints, openPoints[i])
		}
	}
	return filteredPoints
}

func pickRandomPoint(openPoints []*pb.Point) *pb.Point {
	if len(openPoints) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(openPoints))

	return openPoints[randIndex]
}

func getUnoccupiedPoints(width, height int32, food []*pb.Point, snakes []*pb.Snake) []*pb.Point {
	occupiedPoints := getUniqOccupiedPoints(food, snakes)

	numCandidatePoints := (width * height) - int32(len(occupiedPoints))
	if numCandidatePoints <= 0 {
		return []*pb.Point{}
	}

	candidatePoints := make([]*pb.Point, 0, numCandidatePoints)

	for x := int32(0); x < width; x++ {
		for y := int32(0); y < height; y++ {
			p := &pb.Point{X: x, Y: y}
			match := false

			for _, o := range occupiedPoints {
				if o.Equal(p) {
					match = true
					break
				}
			}

			if !match {
				candidatePoints = append(candidatePoints, p)
			}
		}
	}

	return candidatePoints
}

func getUniqOccupiedPoints(food []*pb.Point, snakes []*pb.Snake) []*pb.Point {
	occupiedPoints := []*pb.Point{}
	for _, f := range food {
		candidate := true
		for _, o := range occupiedPoints {
			if o.Equal(f) {
				candidate = false
				break
			}
		}
		if candidate {
			occupiedPoints = append(occupiedPoints, f)
		}
	}

	for _, s := range snakes {
		for _, b := range s.Body {
			candidate := true
			for _, o := range occupiedPoints {
				if o.Equal(b) {
					candidate = false
					break
				}
			}
			if candidate {
				occupiedPoints = append(occupiedPoints, b)
			}
		}
	}

	return occupiedPoints
}

func updateSnakes(game *pb.Game, frame *pb.GameFrame, moves []*SnakeUpdate) {
	for _, update := range moves {
		update.Snake.Latency = fmt.Sprint(int64(update.Latency) / 1e6)
		if update.Err != nil {
			log.WithFields(log.Fields{
				"GameID":  game.ID,
				"Error":   update.Err,
				"SnakeID": update.Snake.ID,
				"Name":    update.Snake.Name,
				"Turn":    frame.Turn,
			}).Info("Default move")
			update.Snake.DefaultMove()
		} else {
			log.WithFields(log.Fields{
				"GameID":  game.ID,
				"SnakeID": update.Snake.ID,
				"Name":    update.Snake.Name,
				"Turn":    frame.Turn,
				"Move":    update.Move,
			}).Info("Non-flip move")
			update.Snake.Move(update.Move)
		}
		if checkForBackflip(update.Snake) {
			log.WithFields(log.Fields{
				"GameID":  game.ID,
				"SnakeID": update.Snake.ID,
				"Name":    update.Snake.Name,
				"Turn":    frame.Turn,
			}).Info("Has flipped")
			update.Snake.Flip()
		}
		if len(update.Snake.Body) != 0 {
			update.Snake.Body = update.Snake.Body[:len(update.Snake.Body)-1]
		}
	}
}

func checkForBackflip(s *pb.Snake) bool {
	if s.Head() == nil || len(s.Body) < 2 {
		return false
	}
	return s.Head().Equal(s.Body[1])
}
