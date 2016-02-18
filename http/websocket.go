package http

import (
	"net/http"
	"fmt"
	"sync"
	"log"
	"encoding/binary"
	"time"
	"bytes"
	"io/ioutil"

	"github.com/pbergman/maze/builder"
	"github.com/gorilla/websocket"
)

var websockets *Websockets

func init() {
	http.HandleFunc("/ws", handleWebSocketRequest)
	websockets = &Websockets{list: make([]*websocket.Conn, 0)}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Websockets struct {
	list []*websocket.Conn
	lock sync.Mutex
}

func (w *Websockets) Add(ws *websocket.Conn) {
	w.lock.Lock()
	w.list = append(w.list, ws)
	w.lock.Unlock()
}

func (w *Websockets) Remove(ws *websocket.Conn) {
	w.lock.Lock()
	defer w.lock.Unlock()
	for i, s := range w.list {
		if s == ws {
			w.list = append(w.list[:i], w.list[i+1:]...)
			break
		}
	}
}

func (w *Websockets) Broadcast(mtype byte,v []byte) {
	w.lock.Lock()
	defer w.lock.Unlock()

	for _, ws:= range w.list {
		m := []byte{mtype}
		ws.WriteMessage(websocket.BinaryMessage, append(m, v...))
	}
}

func handleWebSocketRequest(w http.ResponseWriter, r *http.Request) {
		// https://devcenter.heroku.com/articles/go-websockets
		conn, err := upgrader.Upgrade(w, r, nil)
		checkHttpError(err, w)
		websockets.Add(conn)

		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		for {
			if mt, data, err := conn.ReadMessage(); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					checkHttpError(err, w)
				} else {
					log.Println(err)
				}
				websockets.Remove(conn)
				break

			} else {
				switch mt {
				case websocket.BinaryMessage:
					switch (data[0]) {
					case 1:     // new maze
						height := int(binary.BigEndian.Uint16(data[1:3]));
						width :=  int(binary.BigEndian.Uint16(data[3:5]));
						ratio :=  uint(binary.BigEndian.Uint16(data[5:7]));
						id := time.Now().Unix()
						log.Printf("New Maze(%d): %dx%d [ratio:%d][path:%d,%d,%d][wall:%d,%d,%d]",id, height, width, ratio, data[7],  data[8],  data[9], data[10], data[11], data[12])
						maze := builder.NewMazeImageBuilder(height, width)
						maze.SetRatio(ratio)
						maze.SetPathColor(data[7],  data[8],  data[9])
						maze.SetWallColor(data[10], data[11], data[12])
						matrix, _ := maze.GetMatrix()
						mazes[id] = matrix
						websockets.Broadcast(1, getTemplateList(w))
					case 2:     // new update list
						m := []byte{1}
						err := conn.WriteMessage(websocket.BinaryMessage, append(m, getTemplateList(w)...))
						checkHttpError(err, w)
					case 3:    // get maze
fmt.Println(int64(binary.BigEndian.Uint32(data[1:])))
						m := mazes[int64(binary.BigEndian.Uint32(data[1:]))]
						r := m.I.GetRatio()
						b := make([]byte, 2)
						binary.BigEndian.PutUint16(b, uint16(r))

						p := append(b, m.I.GetWallColor()...)

						b = make([]byte, 2)
						binary.BigEndian.PutUint16(b, uint16(len(m.M)))
						p = append(p, b...)
						b = make([]byte, 2)
						binary.BigEndian.PutUint16(b, uint16(len(m.M[0])))
						p = append(p, b...)


						for y := 0; y < len(m.M)*int(r); y += int(r) {
							for x := 0; x < len(m.M[y/int(r)])*int(r); x += int(r) {
								t :=  m.M[y/int(r)][x/int(r)]
								if  builder.PATH != (builder.PATH & t) {
									p = append(p, byte(x))
									p = append(p, byte(y))
								}
							}
						}
						fmt.Println(p)

						err := conn.WriteMessage(websocket.BinaryMessage, append([]byte{2}, p...))
						checkHttpError(err, w)

					}
				default:
					fmt.Println(mt)
				}
			}
		}

		conn.WriteMessage(websocket.CloseMessage, []byte{})
		websockets.Remove(conn)
}

func getTemplateList(w http.ResponseWriter) []byte {
	buffer := new(bytes.Buffer)
	err := templates.ExecuteTemplate(buffer, "list-group", struct {Mazes map[int64]*builder.MazeImageMatrix}{mazes})
	checkHttpError(err, w)
	read, err := ioutil.ReadAll(buffer)
	checkHttpError(err, w)
	return read
}