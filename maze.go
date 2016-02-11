package main

import (
	"fmt"
	"github.com/pbergman/maze/builder"
	"github.com/pbergman/maze/solver"
	"os"
	"flag"
	"time"
	"strconv"
	"log"
	"net/http"
	"html/template"
	"regexp"
	"image/gif"
)


func main() {


	h := flag.Int("h", 10, "Height of maze")
	w := flag.Int("w",  10, "Width of maze")
	s := flag.Uint("s",  1, "Pixel scale ratio")
	server := flag.Bool("S",  false, "Start Server")
	mf := flag.String("f",  "maze" + strconv.Itoa(int(time.Now().Unix())) + ".gif", "File to write maze to")
	ms := flag.String("sf",  "solved" + strconv.Itoa(int(time.Now().Unix())) + ".gif", "File to write solved maze to")

	flag.Parse()

	if !*server {
		log.Printf("Building maze with ratio %d, width %d, height: %d", *s, *w, *h)
		maze := builder.NewMazeImageBuilder(*h, *w)
		maze.SetRatio(*s)
		log.Printf("Getting matrix and image: %s", maze)
		matrix, err := maze.GetMatrix()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(matrix)
		log.Print("Solving maze")
		start := time.Now()
		walker := solver.NewWalker(matrix)
		walker.Solve()
		log.Printf("Done %s", time.Now().Sub(start))
		fmt.Println(matrix)
		log.Printf("Saving maze: %s", *mf)
		f, _ := os.Create(*mf)
		defer f.Close()
		matrix.ToFile(f)
		log.Printf("Saving solved maze: %s", *ms)
		o, _ := os.Create(*ms)
		defer o.Close()
		walker.ToFile(o)
	} else {

		mazes := make(map[int64]*builder.MazeImageMatrix, 0)
		show  := regexp.MustCompile(`^/show/(\d+)$`)
		solved  := regexp.MustCompile(`^/solved/(\d+)$`)

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

			if r.Method == "POST" {
				r.ParseForm()
				h, _ := strconv.Atoi(r.Form.Get("height"))
				w, _ := strconv.Atoi(r.Form.Get("width"))
				a, _ := strconv.Atoi(r.Form.Get("ratio"))
				br, _ := strconv.Atoi(r.Form.Get("br"))
				bg, _ := strconv.Atoi(r.Form.Get("bg"))
				bb, _ := strconv.Atoi(r.Form.Get("bb"))
				pr, _ := strconv.Atoi(r.Form.Get("pr"))
				pg, _ := strconv.Atoi(r.Form.Get("pg"))
				pb, _ := strconv.Atoi(r.Form.Get("pb"))

				maze := builder.NewMazeImageBuilder(h,w)
				maze.SetRatio(uint(a))
				maze.SetPathColor(byte(pr), byte(pg), byte(pb))
				maze.SetWallColor(byte(br), byte(bg), byte(bb))
				matrix, _ := maze.GetMatrix()
				mazes[time.Now().UnixNano()] = matrix
			}

			if r.Method == "GET" && show.MatchString(r.URL.String()) {

				id, _ := strconv.Atoi(show.FindStringSubmatch(r.URL.String())[1])
				if m, o := mazes[int64(id)]; o {
					gif.Encode(w, m.DrawImage(), &gif.Options{NumColors: 256})
				}
			}

			if r.Method == "GET" && solved.MatchString(r.URL.String()) {
				id, _ := strconv.Atoi(solved.FindStringSubmatch(r.URL.String())[1])
				if m, o := mazes[int64(id)]; o {
					walker := solver.NewWalker(m)
					walker.Solve()
					gif.Encode(w, walker.DrawImage(), &gif.Options{NumColors: 256})
				}
			}

			t, _ := template.ParseFiles("template/base.html")
			t.Execute(w, struct{Mazes map[int64]*builder.MazeImageMatrix}{mazes})

		})
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

}
