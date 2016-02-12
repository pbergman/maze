package builder

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io"
	"os"
)

type MatrixToken uint32

const (
	WALL MatrixToken = 1 << iota
	PATH
	BORDER
	START
	END
)

type MazeImageMatrix struct {
	M [][]MatrixToken
	I *MazeImageBuilder
}

func (i MazeImageMatrix) Has(x, y int, t MatrixToken) bool {
	return t == (t & i.M[y][x])
}

// String prints the the matrix to the stdout in visula way
func (i MazeImageMatrix) String() string {
	buff := new(bytes.Buffer)
	for _, data := range i.M {
		for _, token := range data {
			switch true {
			case WALL == (WALL & token):
				buff.Write([]byte{'#'})
			case PATH == (PATH & token), BORDER == (BORDER & token):
				buff.Write([]byte{' '})
			}
		}
		buff.Write([]byte{'\n'})
	}
	return string(buff.Bytes())
}

// isWall check if the given colors is matching the config wallcolors
func (i *MazeImageMatrix) isWall(r, g, b uint8) bool {
	return i.I.wall_color.R == r && i.I.wall_color.G == g && i.I.wall_color.B == b
}

// NewMazeImageMatrix creates a image matrix base on fetched body
func NewMazeImageMatrix(r io.Reader, m *MazeImageBuilder) (*MazeImageMatrix, error) {
	MazeImageMatrix := &MazeImageMatrix{I: m}
	img, err := gif.Decode(r)
	if err != nil {
		return nil, err
	}
	rect := img.Bounds()
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rect, img, rect.Min, draw.Src)
	reader := bytes.NewReader(rgba.Pix)
	matrix := make([][]MatrixToken, rect.Max.Y)
	for y := 0; y < rect.Max.Y; y++ {
		for x := 0; x < rect.Max.X; x++ {
			if len(matrix[y]) == 0 {
				matrix[y] = make([]MatrixToken, rect.Max.X)
			}
			part := make([]byte, 4)
			reader.Read(part)
			if y == 0 || x == 0 {
				matrix[y][x] = BORDER
			} else {
				if MazeImageMatrix.isWall(part[0], part[1], part[2]) {
					matrix[y][x] = WALL
				} else {
					matrix[y][x] = PATH
				}
			}
		}
	}
	MazeImageMatrix.M = matrix
	return MazeImageMatrix, nil
}

// ToFile will push the drawing created with DrawImage to the given fd
func (i MazeImageMatrix) ToFile(file *os.File) error {
	return gif.Encode(file, i.DrawImage(), &gif.Options{NumColors: 256})
}

// DrawImage draws a new image beased on matrix and config ration
func (i MazeImageMatrix) DrawImage() draw.Image {
	rect := image.Rect(0, 0, len(i.M[0])*int(i.I.ratio), len(i.M)*int(i.I.ratio))
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)
	for y := 0; y < len(i.M)*int(i.I.ratio); y += int(i.I.ratio) {
		for x := 0; x < len(i.M[y/int(i.I.ratio)])*int(i.I.ratio); x += int(i.I.ratio) {
			switch i.M[y/int(i.I.ratio)][x/int(i.I.ratio)] {
			case WALL:
				draw.Draw(rgba, image.Rect(x, y, x+int(i.I.ratio), y+int(i.I.ratio)), &image.Uniform{i.I.wall_color}, image.ZP, draw.Src)
			case PATH, BORDER:
				draw.Draw(rgba, image.Rect(x, y, x+int(i.I.ratio), y+int(i.I.ratio)), &image.Uniform{i.I.path_color}, image.ZP, draw.Src)
			}
		}
	}
	return rgba
}
