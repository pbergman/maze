package solver

import (
	"github.com/pbergman/maze/builder"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
)

type Direction uint16

func (d Direction) String() string {
	switch d {
	case LEFT:
		return "LEFT"
	case UP:
		return "UP"
	case RIGHT:
		return "RIGHT"
	case DOWN:
		return "DOWN"
	default:
		return "UNKNOWN"
	}
}

const (
	LEFT Direction = 1 << iota
	UP
	RIGHT
	DOWN
)

type Walker struct {
	s  Position                 // start point
	e  Position                 // end point
	b  image.Rectangle          // maze bounds, to set borders
	m  *builder.MazeImageMatrix //
	ls int                      // last section id
}

// NewWalker initialize walker and determine the start/end points
func NewWalker(m *builder.MazeImageMatrix) *Walker {

	w := &Walker{
		b: image.Rect(1, 1, len(m.M[0])-2, len(m.M)-2),
		m: m,
	}

	done := false
	check := func(p Position, walker *Walker) bool {
		if walker.m.Has(p.x, p.y, builder.PATH) {
			if walker.s.x == 0 && walker.s.y == 0 {
				walker.s.y, walker.s.x = p.y, p.x
				walker.m.M[p.y][p.x] |= builder.START
				return false
			}
			if walker.e.x == 0 && walker.e.y == 0 {
				walker.e.y, walker.e.x = p.y, p.x
				walker.m.M[p.y][p.x] |= builder.END
				return true
			}
		}
		return false
	}

	pos := Position{1, 1}

	for w.right(&pos) {
		done = check(pos, w)
	}

	for !done && w.down(&pos) {
		done = check(pos, w)
	}

	for !done && w.left(&pos) {
		done = check(pos, w)
	}

	for !done && w.up(&pos) {
		done = check(pos, w)
	}

	// add starting point as visited
	w.m.M[w.s.y][w.s.x] |= builder.VISITED

	return w
}

// peekAround will return array with available direction based on current position
func (w *Walker) peekAround(t TraceablePosition) []Direction {

	directions := make([]Direction, 0)
	left, right, up, down := [2]int{t.x - 1, t.y}, [2]int{t.x + 1, t.y}, [2]int{t.x, t.y - 1}, [2]int{t.x, t.y + 1}

	if left[0] >= w.b.Min.X && w.m.Has(left[0], left[1], builder.PATH) && !t.HasVisited(left[0], left[1]) {
		directions = append(directions, LEFT)
	}

	if right[0] <= w.b.Max.X && w.m.Has(right[0], right[1], builder.PATH) && !t.HasVisited(right[0], right[1]) {
		directions = append(directions, RIGHT)
	}

	if up[1] >= w.b.Min.Y && w.m.Has(up[0], up[1], builder.PATH) && !t.HasVisited(up[0], up[1]) {
		directions = append(directions, UP)
	}

	if down[1] <= w.b.Max.Y && w.m.Has(down[0], down[1], builder.PATH) && !t.HasVisited(down[0], down[1]) {
		directions = append(directions, DOWN)
	}

	return directions
}

func (w *Walker) ToFile(file *os.File) error {
	return gif.Encode(file, w.DrawImage(), &gif.Options{NumColors: 256})
}

// DrawImage draws a new image beased on matrix and config ration
func (w Walker) DrawImage() draw.Image {
	maze := w.m.DrawImage()
	for y := 0; y < len(w.m.M)*int(w.m.I.GetRatio()); y += int(w.m.I.GetRatio()) {
		for x := 0; x < len(w.m.M[y/int(w.m.I.GetRatio())])*int(w.m.I.GetRatio()); x += int(w.m.I.GetRatio()) {
			token := w.m.M[y/int(w.m.I.GetRatio())][x/int(w.m.I.GetRatio())]
			if builder.VISITED == (builder.VISITED & token) {
				draw.Draw(maze, image.Rect(x, y, x+int(w.m.I.GetRatio()), y+int(w.m.I.GetRatio())), &image.Uniform{color.RGBA{133, 133, 133, 150}}, image.ZP, draw.Src)
			}
			if builder.ROUTE == (builder.ROUTE & token) {
				draw.Draw(maze, image.Rect(x, y, x+int(w.m.I.GetRatio()), y+int(w.m.I.GetRatio())), &image.Uniform{color.RGBA{255, 0, 0, 150}}, image.ZP, draw.Src)
			}
		}
	}
	return maze
}

// Will try to solve give maze
func (w *Walker) Solve() {

	walker := NewTraceablePosition(w.s.x, w.s.y)
	walker.AddTrace(w.s.x, w.s.y)

	for {

		if builder.END == (builder.END & w.m.M[walker.y][walker.x]) {
			walker.AddTrace(walker.x, walker.y)
			break
		}

		if peek := w.peekAround(*walker); len(peek) == 0 {
			walker.GoBack()
		} else {

			if len(peek) > 1 {
				walker.AddTraceSection(walker.x, walker.y, len(peek)-1)
			} else {
				walker.AddTrace(walker.x, walker.y)
			}

			switch peek[0] {
			case RIGHT:
				w.right(&walker.Position)
			case DOWN:
				w.down(&walker.Position)
			case UP:
				w.up(&walker.Position)
			case LEFT:
				w.left(&walker.Position)
			}

			w.m.M[walker.y][walker.x] |= builder.VISITED
		}
	}

	for _, d := range walker.t {
		if OK == (OK & d.t) {
			w.m.M[d.y][d.x] |= builder.ROUTE
		}
	}

}

func (m *Walker) left(p *Position) bool {
	if p.x > m.b.Min.X {
		p.x--
		return true
	}
	return false
}

func (m *Walker) right(p *Position) bool {
	if p.x < m.b.Max.X {
		p.x++
		return true
	}
	return false
}

func (m *Walker) up(p *Position) bool {
	if p.y > m.b.Min.Y {
		p.y--
		return true
	}
	return false
}

func (m *Walker) down(p *Position) bool {
	if p.y < m.b.Max.Y {
		p.y++
		return true
	}
	return false
}
