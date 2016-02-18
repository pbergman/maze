package http

import (
	"net/http"
	"github.com/pbergman/maze/builder"
)

var mazes map[int64]*builder.MazeImageMatrix

func init() {
	http.HandleFunc("/", handelWebRequest)
	mazes = make(map[int64]*builder.MazeImageMatrix, 0)
}

func handelWebRequest(w http.ResponseWriter, r *http.Request) {
	templates.Execute(w, struct {Mazes map[int64]*builder.MazeImageMatrix}{mazes})
}

func checkHttpError(err error, w http.ResponseWriter) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}