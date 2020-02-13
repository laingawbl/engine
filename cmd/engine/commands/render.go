package commands

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/laingawbl/engine/controller/pb"
	"github.com/mattn/go-runewidth"
	termbox "github.com/nsf/termbox-go"
)

const (
	defaultColor = termbox.ColorDefault
	bgColor      = termbox.ColorDefault
	snakeColor   = termbox.ColorGreen
)

func render(game *pb.Game, frame *pb.GameFrame) error {
	if frame == nil {
		return errors.New("received nil frame")
	}
	err := termbox.Clear(defaultColor, defaultColor)
	if err != nil {
		return err
	}

	var (
		_, h   = termbox.Size()
		midY   = h / 2
		left   = 10
		top    = 10
		bottom = midY + (int(game.Height) / 2) + 1
	)

	renderTitle(left, top, int(frame.Turn))
	renderBoard(game, top, bottom, left)
	snakePos := 0
	for _, s := range frame.Snakes {
		renderSnake(left, top, s)

		text := fmt.Sprintf("%s %d/100", s.Name, s.Health)
		if s.Death != nil {
			text = fmt.Sprintf("%s - %s", text, s.Death.Cause)
		}
		tbprint(int(game.Width)+left+5, top+snakePos, defaultColor, defaultColor, text)
		snakePos++
		healthColor := termbox.ColorGreen
		for i := 0; i < 10; i++ {
			if int(s.Health) <= ((i * 10) + 1) {
				healthColor = termbox.ColorRed
			}
			termbox.SetCell(int(game.Width)+left+5+i, top+snakePos, ' ', healthColor, healthColor)
		}
		snakePos += 2
	}
	renderFood(left, top, frame.Food)

	return termbox.Flush()
}

func renderSnake(left, top int, s *pb.Snake) {
	for _, b := range s.Body {
		termbox.SetCell(left+int(b.X), top+int(b.Y)+1, ' ', snakeColor, snakeColor)
	}
}

func renderFood(left, top int, food []*pb.Point) {
	for _, f := range food {
		termbox.SetCell(left+int(f.X), top+int(f.Y)+1, getFoodEmoji(f.X, f.Y), defaultColor, bgColor)
	}
}

var foods = map[string]rune{}

func getFoodEmoji(x, y int32) rune {
	key := fmt.Sprintf("(%d, %d)", x, y)
	r, ok := foods[key]
	if !ok {
		r = randomFoodEmoji()
		foods[key] = r
	}
	return r
}

func randomFoodEmoji() rune {
	f := []rune{
		'🍒',
		'🍍',
		'🍑',
		'🍇',
		'🍏',
		'🍌',
		'🍫',
		'🍭',
		'🍕',
		'🍩',
		'🍗',
		'🍖',
		'🍬',
		'🍤',
		'🍪',
	}

	return f[rand.Intn(len(f))]
}

func renderBoard(game *pb.Game, top, bottom, left int) {
	for i := top + 1; i < bottom; i++ {
		termbox.SetCell(left-1, i, '│', defaultColor, bgColor)
		termbox.SetCell(left+int(game.Width), i, '│', defaultColor, bgColor)
	}

	termbox.SetCell(left-1, top, '┌', defaultColor, bgColor)
	termbox.SetCell(left-1, bottom, '└', defaultColor, bgColor)
	termbox.SetCell(left+int(game.Width), top, '┐', defaultColor, bgColor)
	termbox.SetCell(left+int(game.Width), bottom, '┘', defaultColor, bgColor)

	fill(left, top, int(game.Width), 1, termbox.Cell{Ch: '─'})
	fill(left, bottom, int(game.Width), 1, termbox.Cell{Ch: '─'})
}

func renderTitle(left, top, turn int) {
	tbprint(left, top-1, defaultColor, defaultColor, fmt.Sprintf("Battlesnake! - Turn %d", turn))
}

func fill(x, y, w, h int, cell termbox.Cell) {
	for ly := 0; ly < h; ly++ {
		for lx := 0; lx < w; lx++ {
			termbox.SetCell(x+lx, y+ly, cell.Ch, cell.Fg, cell.Bg)
		}
	}
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}
