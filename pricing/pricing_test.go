package pricing

import (
	"golang.org/x/net/websocket"
	"log"
	"log/slog"
	"testing"
	"time"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func TestNewQuotes(t *testing.T) {
	//ws, err := Start()
	//if err != nil {
	//	t.Fatalf("Could not start simulation pricing %v", err)
	//}
	//
	//c := NewQuotes([]string{"AAPL"}, slog.Default())
	//
	//for price := range c {
	//	log.Printf("Received price %s", price)
	//}

}

func TestPricingReconnect(t *testing.T) {
	ws, err := Start()
	if err != nil {
		t.Fatalf("START %v", err)
	}
	defer ws.Close()
	slog.Info("Ready to receive messages")
	var incomingEq EquityQuote

	for i := 0; i < 2; i++ {
		//time.Sleep(17 * time.Second)
		log.Printf("Getting message")
		websocket.JSON.Receive(ws, &incomingEq)
		//if err != nil {
		//	log.Fatalf("%v", err)
		//}
		time.Sleep(1 * time.Second)
		log.Printf("Received: %s", incomingEq)
	}
}
