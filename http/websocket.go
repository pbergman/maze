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
	"github.com/pbergman/maze/solver"
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
						log.Printf("New Maze(%d): %dx%d [ratio:%d][wall:%d,%d,%d][path:%d,%d,%d]",id, height, width, ratio, data[7],  data[8],  data[9], data[10], data[11], data[12])
						maze := builder.NewMazeImageBuilder(int(height), int(width))
						maze.SetRatio(uint(ratio))
						maze.SetWallColor(byte(data[7]),  byte(data[8]),  byte(data[9]))
						maze.SetPathColor(byte(data[10]), byte(data[11]), byte(data[12]))
						matrix, _ := maze.GetMatrix()
						mazes[id] = matrix
						websockets.Broadcast(1, getTemplateList(w))
					case 2:     // new update list
						m := []byte{1}
						err := conn.WriteMessage(websocket.BinaryMessage, append(m, getTemplateList(w)...))
						checkHttpError(err, w)
					case 3:    // get maze
						if m, ok := mazes[int64(binary.BigEndian.Uint32(data[1:]))]; ok {
							ratio := m.I.GetRatio()
							buf := new(bytes.Buffer)
							buf.Write([]byte{2}) // type id
							binary.Write(buf, binary.BigEndian, binary.BigEndian.Uint32(data[1:]))
							binary.Write(buf, binary.BigEndian, uint16(ratio))
							binary.Write(buf, binary.BigEndian, uint16(len(m.M) * int(ratio)))
							binary.Write(buf, binary.BigEndian, uint16(len(m.M[0])  * int(ratio)))
							buf.Write(m.I.GetWallColor())
							buf.Write(m.I.GetPathColor())
							for y := 0; y < len(m.M)*int(ratio); y += int(ratio) {
								for x := 0; x < len(m.M[y/int(ratio)])*int(ratio); x += int(ratio) {
									t :=  m.M[y/int(ratio)][x/int(ratio)]
									if  builder.WALL == (builder.WALL & t) {
										binary.Write(buf, binary.BigEndian, uint16(x))
										binary.Write(buf, binary.BigEndian, uint16(y))
									}
								}
							}
							err := conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
							checkHttpError(err, w)

						} else {
							http.Error(w, fmt.Sprintf("No maze exist by id %d", int64(binary.BigEndian.Uint32(data[1:]))), 500)
						}
					}
				case 4:
					if m, ok := mazes[int64(binary.BigEndian.Uint32(data[1:]))]; ok {

						walker := solver.NewWalker(m)
						walker.Solve()

						ratio := m.I.GetRatio()
						buf := new(bytes.Buffer)
						buf.Write([]byte{3}) // type id
						binary.Write(buf, binary.BigEndian, uint16(ratio))

						for _, t := range walker.GetResult().GetTraces() {
							binary.Write(buf, binary.BigEndian, uint16(t.X*ratio))
							binary.Write(buf, binary.BigEndian, uint16(t.Y*ratio))
							binary.Write(buf, binary.BigEndian, uint8(t.T))
						}

						err := conn.WriteMessage(websocket.BinaryMessage, buf.Bytes())
						checkHttpError(err, w)

					} else {
						http.Error(w, fmt.Sprintf("No maze exist by id %d", int64(binary.BigEndian.Uint32(data[1:]))), 500)
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