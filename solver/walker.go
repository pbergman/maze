package solver

import (
	"github.com/pbergman/maze/builder"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
	"bytes"
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
	r  *TraceablePosition       // result
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
	ratio := int(w.m.I.GetRatio())
	for _, t := range w.r.t {
		rect := image.Rect(t.x*ratio, t.y*ratio, t.x*ratio+int(w.m.I.GetRatio()), t.y*ratio+int(w.m.I.GetRatio()))
		switch (true) {
		case OK == (OK & t.t):
			draw.Draw(maze, rect, &image.Uniform{color.RGBA{255, 0, 0, 150}}, image.ZP, draw.Src)
		default:
			draw.Draw(maze, rect, &image.Uniform{color.RGBA{133, 133, 133, 150}}, image.ZP, draw.Src)
		}
	}
	return maze
}

// DrawImage draws a new image beased on matrix and config ration
func (w Walker) CreateAnimationImage(file *os.File) {

	var palette color.Palette = color.Palette{}
	palette = append(palette, color.White)
	palette = append(palette, color.Black)
	palette = append(palette, color.RGBA{133, 133, 133, 150})
	palette = append(palette, color.RGBA{255, 0, 0, 150})

	out := &gif.GIF{}
	ratio := int(w.m.I.GetRatio())
	for _, t := range w.r.t {
		maze := w.m.DrawImage()
		rect := image.Rect(t.x*ratio, t.y*ratio, t.x*ratio+int(w.m.I.GetRatio()), t.y*ratio+int(w.m.I.GetRatio()))
		switch (true) {
		case OK == (OK & t.t):
			draw.Draw(maze, rect, &image.Uniform{color.RGBA{255, 0, 0, 150}}, image.ZP, draw.Src)
		default:
			draw.Draw(maze, rect, &image.Uniform{color.RGBA{133, 133, 133, 150}}, image.ZP, draw.Src)
		}
		pm := image.NewPaletted(maze.Bounds(), palette)
		draw.FloydSteinberg.Draw(pm, maze.Bounds(), maze, image.ZP)
		out.Image = append(out.Image, pm)
		out.Delay = append(out.Delay, 0)
	}
	gif.EncodeAll(file, out)
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
			walker.AddTrace(walker.x, walker.y)
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
		}
	}

	w.r = walker
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

// String prints the the matrix to the stdout in visula way
func (w Walker) String() string {
	buff := new(bytes.Buffer)
	for y, data := range w.m.M {
		for x, token := range data {

			if trace := w.r.GetTrace(x, y); trace != nil {
				switch (true) {
				case OK == (OK & trace.t):
					buff.Write([]byte{'*'})
				case VISITED == (VISITED & trace.t):
					buff.Write([]byte{'.'})
				}

			} else {
				switch true {
				case builder.WALL == (builder.WALL & token):
					buff.Write([]byte{'#'})
				case builder.PATH == (builder.PATH & token), builder.BORDER == (builder.BORDER & token):
					buff.Write([]byte{' '})
				case builder.START == (builder.START & token):
					buff.Write([]byte{'S'})
				case builder.END == (builder.END & token):
					buff.Write([]byte{'E'})
				}
			}

		}
		buff.Write([]byte{'\n'})
	}
	return string(buff.Bytes())
}
