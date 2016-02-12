package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pbergman/maze/builder"
	"github.com/pbergman/maze/solver"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type AppConfig struct {
	width    int
	height   int
	scale    uint
	files    struct{ raw, solved, animation string }
	server   bool
	port     int
	template string
}

var config *AppConfig

func init() {
	config = &AppConfig{}
	flag.IntVar(&config.height, "h", 10, "Height of maze")
	flag.IntVar(&config.width, "w", 10, "Width of maze")
	flag.UintVar(&config.scale, "s", 1, "Pixel scale ratio")
	flag.BoolVar(&config.server, "S", false, "Start web server")
	flag.IntVar(&config.port, "p", 8080, "Port to listen for server")
	flag.StringVar(&config.template, "t", "template/base.html", "Template location for server")
	flag.StringVar(&config.files.raw, "rf", "rmaze."+strconv.Itoa(int(time.Now().Unix()))+".gif", "File name to write maze to")
	flag.StringVar(&config.files.solved, "sf", "smaze."+strconv.Itoa(int(time.Now().Unix()))+".gif", "File name to write solved maze to")
	flag.StringVar(&config.files.animation, "af", "amaze."+strconv.Itoa(int(time.Now().Unix()))+".gif", "File name to write anmation maze to")
	flag.Parse()
}

func main() {

	if !config.server {
		cliApp()
	} else {
		webApp()
	}

}

func cliApp() {
	var wg sync.WaitGroup
	log.Printf("Building maze with ratio %d, width %d, height: %d", config.scale, config.width, config.height)
	maze := builder.NewMazeImageBuilder(config.height, config.width)
	maze.SetRatio(config.scale)
	log.Printf("Getting matrix and image: %s", maze)
	matrix, err := maze.GetMatrix()
	checkError(err)
	fmt.Println(matrix)
	log.Print("Solving maze")
	start := time.Now()
	walker := solver.NewWalker(matrix)
	walker.Solve()
	log.Printf("Done %s", time.Now().Sub(start))
	fmt.Println(walker)
	wg.Add(2)
	go func() {
		log.Printf("Saving maze: %s", config.files.raw)
		f, err := os.Create(config.files.raw)
		checkError(err)
		defer f.Close()
		matrix.ToFile(f)
		wg.Done()
	}()
	go func() {
		log.Printf("Saving solved maze: %s", config.files.solved)
		o, err := os.Create(config.files.solved)
		checkError(err)
		defer o.Close()
		walker.ToFile(o)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		log.Printf("Saving solved animation maze: %s", config.files.animation)
		o, err := os.Create(config.files.animation)
		checkError(err)
		defer o.Close()
		walker.CreateAnimationImage(o)
		wg.Done()
	}()
	wg.Wait()
}

func webApp() {

	mazes := make(map[int64]*builder.MazeImageMatrix, 0)
	show := regexp.MustCompile(`^/show/(\d+)$`)
	solved := regexp.MustCompile(`^/solved/(\d+)$`)
	animation := regexp.MustCompile(`^/animation/(\d+)$`)
	tmpFolder := os.TempDir() + "/mazes"

	os.RemoveAll(tmpFolder)
	os.Mkdir(tmpFolder, os.ModePerm)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "POST":
			r.ParseForm()
			fh, _ := strconv.Atoi(r.Form.Get("height"))
			fw, _ := strconv.Atoi(r.Form.Get("width"))
			fa, _ := strconv.Atoi(r.Form.Get("ratio"))
			br, _ := strconv.Atoi(r.Form.Get("br"))
			bg, _ := strconv.Atoi(r.Form.Get("bg"))
			bb, _ := strconv.Atoi(r.Form.Get("bb"))
			pr, _ := strconv.Atoi(r.Form.Get("pr"))
			pg, _ := strconv.Atoi(r.Form.Get("pg"))
			pb, _ := strconv.Atoi(r.Form.Get("pb"))

			maze := builder.NewMazeImageBuilder(fh, fw)
			maze.SetRatio(uint(fa))
			maze.SetPathColor(byte(pr), byte(pg), byte(pb))
			maze.SetWallColor(byte(br), byte(bg), byte(bb))
			matrix, _ := maze.GetMatrix()
			mazes[time.Now().UnixNano()] = matrix

			keys := make([]string, 0)

			for i, _ := range mazes {
				keys = append(keys, strconv.Itoa(int(i)))
			}

			js, err := json.Marshal(keys)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		case "GET":
			switch true {
			case show.MatchString(r.URL.String()):
				id, _ := strconv.Atoi(show.FindStringSubmatch(r.URL.String())[1])
				if m, o := mazes[int64(id)]; o {
					file := fmt.Sprintf("%s/r%d.gif", tmpFolder, id)
					if _, err := os.Stat(file); os.IsNotExist(err) {
						fd, err := os.Create(file)
						checkHttpError(err, w)
						m.ToFile(fd)
						fd.Close()
					}
					fb, err := ioutil.ReadFile(file)
					checkHttpError(err, w)
					buffer := bytes.NewBuffer(fb)
					w.Header().Set("Content-Type", "image/gif")
					_, err = buffer.WriteTo(w)
					checkHttpError(err, w)
				} else {
					http.NotFound(w, r)
				}
			case solved.MatchString(r.URL.String()):
				id, _ := strconv.Atoi(solved.FindStringSubmatch(r.URL.String())[1])
				if m, o := mazes[int64(id)]; o {
					file := fmt.Sprintf("%s/s%d.gif", tmpFolder, id)
					if _, err := os.Stat(file); os.IsNotExist(err) {
						fd, err := os.Create(file)
						checkHttpError(err, w)
						walker := solver.NewWalker(m)
						walker.Solve()
						walker.ToFile(fd)
						fd.Close()
					}
					fb, err := ioutil.ReadFile(file)
					checkHttpError(err, w)
					buffer := bytes.NewBuffer(fb)
					w.Header().Set("Content-Type", "image/gif")
					_, err = buffer.WriteTo(w)
					checkHttpError(err, w)
				} else {
					http.NotFound(w, r)
				}
			case animation.MatchString(r.URL.String()):
				id, _ := strconv.Atoi(animation.FindStringSubmatch(r.URL.String())[1])
				if m, o := mazes[int64(id)]; o {
					file := fmt.Sprintf("%s/a%d.gif", tmpFolder, id)
					if _, err := os.Stat(file); os.IsNotExist(err) {
						fd, err := os.Create(file)
						checkError(err)
						walker := solver.NewWalker(m)
						walker.Solve()
						walker.CreateAnimationImage(fd)
						fd.Close()
					}
					fb, err := ioutil.ReadFile(file)
					checkHttpError(err, w)
					buffer := bytes.NewBuffer(fb)
					w.Header().Set("Content-Type", "image/gif")
					_, err = buffer.WriteTo(w)
					checkHttpError(err, w)
				} else {
					http.NotFound(w, r)
				}
			default:
				t, _ := template.ParseFiles("template/base.html")
				t.Execute(w, struct {
					Mazes map[int64]*builder.MazeImageMatrix
				}{mazes})

			}
		}
	})

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.port), nil))
}

func checkHttpError(err error, w http.ResponseWriter) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
