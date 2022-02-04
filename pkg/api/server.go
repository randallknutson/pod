package api

import (
    "fmt"
    "log"
    "net/http"

    "github.com/gorilla/websocket"
    "github.com/avereha/pod/pkg/pod"
)

type Server struct {
  http.Handler

	pod       *pod.Pod
  conn      *websocket.Conn
}

func New(pod *pod.Pod) *Server {

  ret := &Server{
    pod:   pod,
  }

  return ret
}

func (s *Server) Start() {
  fmt.Println("Pod simulator web api listening on :8080")
  s.setupRoutes()
  fmt.Println("Setting StateHook")
  s.pod.SetStateHook(func(state []byte) {
    s.handleNewState(state)
  })
  http.ListenAndServe(":8080", nil)
}

func (s *Server) handleNewState(state []byte) {
  fmt.Println("writing to websocket")
  if s.conn != nil {
    if err := s.conn.WriteMessage(websocket.TextMessage, state); err != nil {
      log.Println(err)
      return
    }
  }
}

func (s *Server) setupRoutes() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "This is an API to the pod simulator intended to be used with a separate web client.")
  })
  http.Handle("/ws", s)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  fmt.Println(r.Host)

  // upgrade this connection to a WebSocket
  // connection
  ws, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Println(err)
  }

  s.conn = ws
  // listen indefinitely for new messages coming
  // through on our WebSocket connection

  s.reader(ws)
}

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
func (s *Server) reader(conn *websocket.Conn) {
  for {
    // Send current state initially
    state, err := s.pod.GetPodStateJson()
    if err := conn.WriteMessage(websocket.TextMessage, state); err != nil {
      log.Println(err)
      return
    }

    _, p, err := conn.ReadMessage()
    if err != nil {
      log.Println(err)
      return
    }
    fmt.Println("Received: " + string(p))
  }
}


// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
  ReadBufferSize:  1024,
  WriteBufferSize: 1024,

  // We'll need to check the origin of our connection
  // this will allow us to make requests from our React
  // development server to here.
  // For now, we'll do no checking and just allow any connection
  CheckOrigin: func(r *http.Request) bool { return true },
}
