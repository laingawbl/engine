package rules

import "github.com/laingawbl/engine/controller/pb"

type deathUpdate struct {
	Snake *pb.Snake
	Death *pb.Death
}

// checkForDeath looks through the snakes with the updated coords and checks to see if any have died
// possible death options are starvation (health has reached 0), wall collision, snake body collision
// snake head collision (other snake is same size or greater)
func checkForDeath(width, height int32, frame *pb.GameFrame) []deathUpdate {
	updates := []deathUpdate{}
	for _, s := range frame.AliveSnakes() {
		if deathByHealth(s.Health) {
			updates = append(updates, deathUpdate{
				Snake: s,
				Death: &pb.Death{
					Turn:  frame.Turn,
					Cause: DeathCauseStarvation,
				},
			})
			continue
		}
		head := s.Head()
		if head == nil {
			continue
		}
		if deathByOutOfBounds(head, width, height) {
			updates = append(updates, deathUpdate{
				Snake: s,
				Death: &pb.Death{
					Turn:  frame.Turn,
					Cause: DeathCauseWallCollision,
				},
			})
			continue
		}

		for _, other := range frame.AliveSnakes() {
			for i, b := range s.Body {
				if i == 0 {
					continue
				}
				// check if other snake's head intersects this snake's body
				if deathByBodyCollision(other.Head(), b) {
					var cause string
					if s.ID == other.ID {
						cause = DeathCauseSnakeSelfCollision
					} else {
						cause = DeathCauseSnakeCollision
					}
					updates = append(updates, deathUpdate{
						Snake: s,
						Death: &pb.Death{
							Turn:  frame.Turn,
							Cause: cause,
						},
					})
					break
				}
			}
		}
	}
	return updates
}

func deathByHealth(health int32) bool {
	return health <= 0
}

func deathByBodyCollision(head, body *pb.Point) bool {
	return head.Equal(body)
}

func deathByOutOfBounds(head *pb.Point, width, height int32) bool {
	return (head.X < 0) || (head.X >= width) || (head.Y < 0) || (head.Y >= height)
}
