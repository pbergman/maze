package solver

type WalkToken uint16

const (
	MULTI WalkToken = 1 << iota
	OK
	VISITED
)

type Position struct {
	x int
	y int
}

type Trace struct {
	X  int       // x position of trace
	Y  int       // y position of trace
	T  WalkToken // value of position
	tr int       // count of directions
	s  int       // section id
}

type TraceablePosition struct {
	Position
	t []Trace // stack of traces
	s []int   // stack of sections
}

func NewTraceablePosition(x, y int) *TraceablePosition {
	return &TraceablePosition{Position{x, y}, make([]Trace, 0), make([]int, 0)}
}

// GoBack will go back to last multi section position and updates the tokens
func (t *TraceablePosition) GoBack() {
	var index int

	// search last multi section position with tries left
	for i, wt := range t.t {
		if MULTI == (MULTI&wt.T) && OK == (OK&wt.T) && wt.tr > 0 {
			index = i
		}
	}

	inSlice := func(i int, s []int) bool {
		for _, l := range s {
			if l == i {
				return true
			}
		}
		return false
	}

	for o, i := range t.s {
		if i == index {
			for i := 0; i < len(t.t); i++ {
				// remove all ok tokens
				if inSlice(t.t[i].s, t.s[o:]) {
					t.t[i].T ^= OK
				}
			}
			t.s = t.s[:o]
			break
		}
	}

	// subtracting counter trying new route
	t.t[index].tr--
	// update position with last section position
	t.y, t.x = t.t[index].Y, t.t[index].X
}

func (t *TraceablePosition) AddTrace(x, y int) {

	if len(t.s) > 0 {
		t.t = append(t.t, Trace{x, y, OK | VISITED, 0, t.s[len(t.s)-1]})
	} else {
		t.t = append(t.t, Trace{x, y, OK | VISITED, 0, 0})
	}

}

func (t *TraceablePosition) AddTraceSection(x, y int, count int) {
	t.s = append(t.s, len(t.t))
	t.t = append(t.t, Trace{x, y, OK | VISITED | MULTI, count, t.s[len(t.s)-1]})
}

// HasVisited will check if given cordinates are in the trace stack
func (t *TraceablePosition) HasVisited(x, y int) bool {
	for _, wt := range t.t {
		if wt.X == x && wt.Y == y {
			return true
		}
	}
	return false
}

func (t *TraceablePosition) GetTraces() []Trace {
	return t.t;
}

func (t *TraceablePosition) GetTrace(x, y int) *Trace {
	for _, wt := range t.t {
		if wt.X == x && wt.Y == y {
			return &wt
		}
	}
	return nil
}
