package rules

import "github.com/laingawbl/engine/controller/pb"

// CheckForGameOver checks if the game has ended. End condition is dependent on game mode.
func CheckForGameOver(mode GameMode, gt *pb.GameFrame) bool {
	aliveSnakes := gt.AliveSnakes()
	if mode == GameModeSinglePlayer {
		return len(aliveSnakes) == 0
	}
	return len(aliveSnakes) == 1 || len(aliveSnakes) == 0
}
