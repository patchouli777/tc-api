package presence

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/lib/sl"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

type Server struct {
	log              *slog.Logger
	ChannelToClients map[Channel]map[*websocket.Conn]bool
	register         chan Conn
	unregister       chan Conn
	broadcast        chan []byte
	mu               sync.Mutex
	vs               *ViewerStore
}

func NewServer(log *slog.Logger, vs *ViewerStore) *Server {
	return &Server{
		log:              log,
		vs:               vs,
		ChannelToClients: make(map[Channel]map[*websocket.Conn]bool),
		register:         make(chan Conn),
		unregister:       make(chan Conn),
		broadcast:        make(chan []byte),
	}
}

func NewHandler(s *Server) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		s.log.Info("new connection")
		defer ws.Close()

		var initMsg InitMessage
		err := websocket.JSON.Receive(ws, &initMsg)
		if err != nil {
			s.log.Error("error receiving init message", sl.Err(err))
			return
		}

		s.register <- Conn{Channel: Channel(initMsg.Channel), Conn: ws}

		defer func() {
			s.unregister <- Conn{Channel: Channel(initMsg.Channel), Conn: ws}
		}()

		for {
			var msg string
			err := websocket.Message.Receive(ws, &msg)
			if err != nil {
				s.log.Error("error receiving message", sl.Err(err))
				break
			}
			s.broadcast <- []byte(msg)
		}
	}
}

func (s *Server) Run(ctx context.Context, instanceID uuid.UUID, timeout time.Duration) {
	go s.StoreViewersUpdate(ctx, instanceID, timeout)

	for {
		select {
		case client := <-s.register:
			if s.ChannelToClients[client.Channel] == nil {
				s.ChannelToClients[client.Channel] = make(map[*websocket.Conn]bool)
			}
			clients := s.ChannelToClients[client.Channel]

			s.mu.Lock()
			clients[client.Conn] = true
			s.mu.Unlock()
			s.log.Info("Client registered")
		case client := <-s.unregister:
			clients := s.ChannelToClients[client.Channel]

			s.mu.Lock()
			if _, ok := clients[client.Conn]; ok {
				delete(clients, client.Conn)
				client.Conn.Close()
				s.log.Info("Client unregistered")
			}
			s.mu.Unlock()
		case <-ctx.Done():
			s.log.Info("shutting down ws server")
			// TODO: impl shutdown
			return
		}
	}
}

// TODO: это заглушка
func (s *Server) StoreViewersUpdate(
	ctx context.Context,
	instanceID uuid.UUID,
	timeout time.Duration) {
	entries, err := os.ReadDir("./static/livestreamthumbs")
	if err != nil {
		s.log.Error("error loading livestreamthumbs while updating livestreams", sl.Err(err))
	}

	thumbnails := make([]string, len(entries))
	for i, entry := range entries {
		thumbnails[i] = "livestreamthumbs\\" + entry.Name()
	}

	for {
		timeoutCtx, cancelTimeout := context.WithCancel(ctx)
		s.log.Info("livestreams update started")

		for channel, viewers := range s.ChannelToClients {
			s.log.Info("publishing update", sl.Str("channel", string(channel)))

			thumbnailId := rand.Intn(len(thumbnails))
			thumbnail := thumbnails[thumbnailId]
			channel := string(channel)
			viewers := len(viewers)

			if err := s.vs.AddInstanceStat(timeoutCtx, channel, instanceID.String(), viewers, thumbnail); err != nil {
				s.log.Error("error setting viewers",
					sl.Str("channel", channel),
					sl.Str("instanceID", instanceID.String()),
					sl.Err(err))
				continue
			}

			s.log.Info("publishing channel", sl.Str("channel", channel))
			if err := s.vs.PublishChannel(ctx, channel); err != nil {
				s.log.Error("publishing channel",
					sl.Str("channel", channel),
					sl.Err(err))
			}

			// if err = u.lsr.UpdateThumbnail(timeoutCtx, ls.User.Name, thumbnails[thumbnailId]); err != nil {
			// 	s.log.Error("error updating thumbnail",
			// 		sl.Err(err),
			// 		sl.Str("user", ls.User.Name),
			// 		sl.Str("category name", ls.Category.Name),
			// 		sl.Str("category link", ls.Category.Link))
			// }
		}

		select {
		case <-ctx.Done():
			s.log.Info("livestreams update ended")
			cancelTimeout()
			return
		case <-time.After(timeout):
			cancelTimeout()
		}
		s.log.Info(fmt.Sprintf("livestreams data updated. next in %s", timeout.String()))
	}
}
