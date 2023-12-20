package socket

import (
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	log "github.com/shyunku-libraries/go-logger"
	"os"
)

var (
	SocketManager = NewSocketManager()
)

func UseSocket(r *gin.Engine) {
	io := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			//&polling.Transport{
			//	Client: &http.Client{
			//		Timeout: time.Minute,
			//	},
			//},
			&websocket.Transport{},
		},
	})
	SocketManager.Io = io
	userHandlers(io)

	go func() {
		err := io.Serve()
		if err != nil {
			log.Error("socket.io server error: ", err)
			log.Fatal(err)
			os.Exit(1)
		}
		log.Infof("socket.io server started")
	}()

	r.GET("/socket.io/", gin.WrapH(io))
	r.POST("/socket.io/", func(c *gin.Context) {
		io.ServeHTTP(c.Writer, c.Request)
	})
}

func userHandlers(io *socketio.Server) {
	io.OnConnect("/", func(s socketio.Conn) error {
		log.Infof("Socket connected: %v", s.ID())
		SocketManager.AddUser(nil, s)
		return nil
	})

	io.OnError("/", func(s socketio.Conn, e error) {
		log.Warnf("Socket error: %v", e)
	})

	io.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Debugf("Socket disconnected: %v", reason)
		SocketManager.RemoveUserByConnId(s.ID())
		s.LeaveAll()
	})

	io.OnEvent("/", EventTest, func(s socketio.Conn, msg string) {
		s.Emit(EventTest, "test")
	})
}
