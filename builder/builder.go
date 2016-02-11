package builder

import (
	"fmt"
	"image/color"
	"net/http"
)

// MazeImage is basic configuration for fetching a maze image
type MazeImageBuilder struct {
	height int
	width  int
	// ration i replacement for path and wall width
	// because we only want square paths/walls this
	// also drops need for factor length. This will
	// also be ignored when fetching image and only
	// used with re-rendering of image
	ratio      uint
	wall_color *color.RGBA
	path_color *color.RGBA
}

func NewMazeImageBuilder(height, width int) *MazeImageBuilder {
	return &MazeImageBuilder{
		height:     height,
		width:      width,
		ratio:      1,
		path_color: &color.RGBA{255, 255, 255, 255},
		wall_color: &color.RGBA{0, 0, 0, 255},
	}
}

// SetPathColor will set path color for maze, default 255,255,255
func (m *MazeImageBuilder) SetPathColor(r, g, b byte) {
	m.path_color.R = r
	m.path_color.G = g
	m.path_color.B = b
}

// SetWallColor will set path color for maze, default 0,0,0
func (m *MazeImageBuilder) SetWallColor(r, g, b byte) {
	m.wall_color.R = r
	m.wall_color.G = g
	m.wall_color.B = b
}

// Will set ration, so ef set to 2 every "pixel block" is 2x2 pixel
func (m *MazeImageBuilder) SetRatio(r uint) {
	m.ratio = r
}

func (m *MazeImageBuilder) GetRatio() uint {
	return m.ratio
}

// String will return url for fetching the maze image
func (m MazeImageBuilder) String() string {
	return fmt.Sprintf(
		"http://www.hereandabove.com/cgi-bin/maze?%d+%d+%d+%d+0+%d+%d+%d+%d+%d+%d",
		m.width,
		m.height,
		1,
		1,
		m.wall_color.R,
		m.wall_color.G,
		m.wall_color.B,
		m.path_color.R,
		m.path_color.G,
		m.path_color.B,
	)
}

// GetMatrix will return and create matrix based on fetched image
func (m *MazeImageBuilder) GetMatrix() (*MazeImageMatrix, error) {
	resp, err := http.Get(m.String())
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	} else {
		if image, err := NewMazeImageMatrix(resp.Body, m); err != nil {
			return nil, err
		} else {
			return image, nil
		}
	}
}
