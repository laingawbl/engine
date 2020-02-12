package pb

// Move the snake 1 space in the specified direction, move does not remove the end point of the snake, that will be done
// after snakes have eaten
func (s *Snake) Move(direction string) {
	h := s.Head()
	if h == nil {
		return
	}
	switch direction {
	case "up":
		s.Body = append([]*Point{
			{X: h.X, Y: h.Y - 1},
		}, s.Body...)
	case "down":
		s.Body = append([]*Point{
			{X: h.X, Y: h.Y + 1},
		}, s.Body...)
	case "left":
		s.Body = append([]*Point{
			{X: h.X - 1, Y: h.Y},
		}, s.Body...)
	case "right":
		s.Body = append([]*Point{
			{X: h.X + 1, Y: h.Y},
		}, s.Body...)
	default:
		s.DefaultMove()
	}
}

// DefaultMove the snake will move 1 space in the direction it was already heading
func (s *Snake) DefaultMove() {
	if len(s.Body) < 2 {
		s.Move("up")
		return
	}
	head := s.Head()
	neck := s.Body[1]

	if head.X == neck.X && head.Y == neck.Y {
		// this is the case when the game starts up and all 3 segments are still on the same point
		s.Move("up")
	} else if head.X == neck.X {
		if head.Y > neck.Y {
			s.Move("down")
		} else {
			s.Move("up")
		}
	} else if head.Y == neck.Y {
		if head.X > neck.X {
			s.Move("right")
		} else {
			s.Move("left")
		}
	}
}

// Flip reverses the Body; if this happens while the head (Body[0]) is on the neck (Body[1]),
// the new tail (former head) is later deleted and the snake will appear to stay in place
func (s *Snake) Flip() {
	for h, t := 0, len(s.Body)-1; h < t; h, t = h+1, t-1 {
		s.Body[h], s.Body[t] = s.Body[t], s.Body[h]
	}
}

// Head returns the first point in the body
func (s *Snake) Head() *Point {
	if len(s.Body) == 0 {
		return nil
	}
	return s.Body[0]
}

func (s *Snake) Tail() *Point {
	if len(s.Body) == 0 {
		return nil
	}
	return s.Body[len(s.Body)-1]
}
