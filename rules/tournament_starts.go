package rules

import "github.com/battlesnakeio/engine/controller/pb"

var (
	// 7x7 board
	smallStarts = make([]*pb.Point, 8)

	// 11x11 board
	mediumStarts = make([]*pb.Point, 8)

	// 19x19 board
	largeStarts = make([]*pb.Point, 8)
)

func init() {
	boards := map[int][]*pb.Point{
		7:  smallStarts,
		11: mediumStarts,
		19: largeStarts,
	}
	for s, starts := range boards {
		size := int32(s)
		center := (size - 1) / 2
		starts[0] = &pb.Point{X: 1, Y: 1}
		starts[1] = &pb.Point{X: size - 2, Y: size - 2}
		starts[2] = &pb.Point{X: 1, Y: size - 2}
		starts[3] = &pb.Point{X: size - 2, Y: 1}
		starts[4] = &pb.Point{X: center, Y: 1}
		starts[5] = &pb.Point{X: size - 2, Y: center}
		starts[6] = &pb.Point{X: center, Y: size - 2}
		starts[7] = &pb.Point{X: 1, Y: center}
	}
}
