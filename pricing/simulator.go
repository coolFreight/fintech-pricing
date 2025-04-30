package pricing

import (
	gSocket "github.com/gorilla/websocket"
	"golang.org/x/net/websocket"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

var upgrader = gSocket.Upgrader{}

func Start() (*websocket.Conn, error) {
	s := httptest.NewServer(http.HandlerFunc(priceSimulation))
	defer s.Close()
	defer log.Println("Server shutdown")

	u := "ws" + strings.TrimPrefix(s.URL, "http")
	slog.Info("Websocket connecting to %s", "url", u)
	ws, err := websocket.Dial(u, "ws", s.URL)
	if err != nil {
		slog.Error("could not connect", "error", err)
		return nil, err
	}
	slog.Info("Websocket connected")
	return ws, nil
}

func priceSimulation(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	if err != nil {
		slog.Error("Websocket dial failed: %v", "error", err)
		return
	}
	defer c.Close()
	quotes := make([]EquityQuote, 0)
	quotes = append(quotes, EquityQuote{BidPrice: 25.66})
	for {
		time.Sleep(5 * time.Second)
		slog.Debug("Sending message")
		err = c.WriteJSON(quotes)
		if err != nil {
			log.Printf("Write JSON failed: %v", err)
		} else {
			slog.Debug("Message sent")
		}
	}
}
