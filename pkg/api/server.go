package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/avereha/pod/pkg/pod"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	http.Handler

	pod  *pod.Pod
	conn *websocket.Conn
}

func New(pod *pod.Pod) *Server {

	ret := &Server{
		pod: pod,
	}

	return ret
}

func (s *Server) Start() {
	fmt.Println("Pod simulator web api listening on :8080")
	s.setupRoutes()
	fmt.Println("Setting Web Message Hook")
	s.pod.SetWebMessageHook(func(msg []byte) {
		s.sendMessage(msg)
	})
	http.ListenAndServe(":8080", nil)
}

func (s *Server) sendMessage(msg []byte) {
	fmt.Println("writing to websocket")
	if s.conn != nil {
		if err := s.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
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
		s.handleCommand(p)
	}
}

func (s *Server) handleCommand(bytes []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(bytes, &msg); err != nil {
		log.Error(err)
		return
	}

	var (
		command string
		value   float64
		ok      bool
	)

	if command, ok = msg["command"].(string); !ok {
		log.Fatal("command is not a string or not in msg")
	}

	switch command {
	case "changeReservoir":
		if value, ok = msg["value"].(float64); !ok {
			log.Fatal("reservoir value is not a number or not in msg")
		}
		s.pod.SetReservoir(float32(value))
	case "setAlerts":
		if value, ok = msg["value"].(float64); !ok {
			log.Fatal("alert value is not a number or not in msg")
		}
		s.pod.SetAlerts(uint8(value))
	case "setFault":
		if value, ok = msg["value"].(float64); !ok {
			log.Fatal("fault value is not a number or not in msg")
		}
		s.pod.SetFault(uint8(value))
	case "setActiveTime":
		if value, ok = msg["value"].(float64); !ok {
			log.Fatal("active time in minutes is not a number or not in msg")
		}
		s.pod.SetActiveTime(int(value))
	case "crashNextCommand":
		var beforeProcessing bool
		if beforeProcessing, ok = msg["beforeProcessing"].(bool); !ok {
			log.Fatal("beforeProcessing is not a bool or not in msg")
		}
		s.pod.CrashNextCommand(beforeProcessing)
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
