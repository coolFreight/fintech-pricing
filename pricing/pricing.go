package pricing

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/coolFreight/fintech-pricing/internal"
	"golang.org/x/net/websocket"
)

type EquityQuote struct {
	MessageType string   `json:"T"`
	Symbol      string   `json:"S"`
	AskExchange string   `json:"ax"`
	AskPrice    float64  `json:"ap"`
	AskSize     int64    `json:"as"`
	BidExchange string   `json:"bx"`
	BidPrice    float64  `json:"bp"`
	BidSize     int64    `json:"bs"`
	Condition   []string `json:"c"`
	Timestamp   string   `json:"t"`
	Tape        string   `json:"z"`
}

type CryptoQuote struct {
	MessageType string  `json:"T"`
	Symbol      string  `json:"S"`
	BidPrice    float64 `json:"bp"`
	BidSize     float64 `json:"bs"`
	AskPrice    float64 `json:"ap"`
	AskSize     float64 `json:"as"`
	Timestamp   string  `json:"t"`
}

type TradeConnect struct {
	Action string   `json:"action"`
	Trades []string `json:"trades"`
}

type PricingConnect struct {
	Action string   `json:"action"`
	Quotes []string `json:"quotes"`
}

type PricingClient struct {
	conn         *websocket.Conn
	logger       *slog.Logger
	EquityQuotes <-chan []EquityQuote
}

func NewPricingClient(conn *websocket.Conn, done <-chan any, tickers []string, logger *slog.Logger) *PricingClient {
	logger.Info(fmt.Sprintf("Subscribing for pricing for tickers : %s ", tickers))
	quotes := PricingConnect{Action: "subscribe", Quotes: tickers}
	err := send(quotes, conn, logger)
	if err != nil {
		logger.Error("Could not subscribe to quotes ", "error", err)
		return nil
	}
	read(conn, logger)
	logger.Info(fmt.Sprintf("Successfully subscribed for quotes %s", tickers))
	quoteChan := make(chan []EquityQuote)
	go func() {
		defer fmt.Println("shutting down quote channel")
		defer close(quoteChan)
		defer conn.Close()
		for {
			var quotes []EquityQuote
			if err := websocket.JSON.Receive(conn, &quotes); err != nil {
				if err != nil {
					logger.Error("Error receiving a pricing quotes", "error", err)
				}
				err = conn.Close()
				if err != nil {
					logger.Error("Error closing connection", "error", err)
				}
				conn, err = reconnect(logger, tickers)
				if err != nil {
					logger.Error("Could not reconnect to pricing", "error", err)
					time.Sleep(30 * time.Second)
				}
			}
			logger.Info(fmt.Sprintf("Received quotes: %v", quotes))
			select {
			case quoteChan <- quotes:
			case <-done:
				return
			}
		}
	}()

	return &PricingClient{
		conn:         conn,
		logger:       logger,
		EquityQuotes: quoteChan,
	}
}

func reconnect(logger *slog.Logger, tickers []string) (*websocket.Conn, error) {
	logger.Info("Reconnecting for pricing")
	conn, err := internal.Connect()
	if err != nil {
		logger.Error("Could not connect for pricing", "error", err)
	} else {
		quotes := PricingConnect{Action: "subscribe", Quotes: tickers}
		err = send(quotes, conn, logger)
		if err != nil {
			logger.Error("Could not subscribe to quotes ", "error", err)
		} else {
			read(conn, logger)
			logger.Info("ReSubscribed to quotes")
		}
	}
	return conn, err
}

func (q EquityQuote) String() string {
	return fmt.Sprintf("EquityQuote"+
		"{Ticker: %s, "+
		"BidPrice: %f, BidSize: %d, BidExchange: %s, AskPrice: %f, AskSize: %d, AskEchange: %s, Time: %s}",
		q.Symbol, q.BidPrice, q.BidSize, q.BidExchange, q.AskPrice, q.AskSize, q.AskExchange, q.Timestamp)
}

func send(data any, ws *websocket.Conn, logger *slog.Logger) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logger.Error(fmt.Sprintf("could not marshal data %s", data), "error", err)
	}
	return websocket.Message.Send(ws, dataBytes)
}

func (pc *PricingClient) Subscribe(tickers []string) error {
	pc.logger.Info("Subscribing for pricing for tickers", slog.Any("tickers", tickers))
	dataBytes, err := json.Marshal(PricingConnect{Action: "subscribe", Quotes: tickers})
	if err != nil {
		pc.logger.Error("could subscribe for tickers", slog.Any("tickers", tickers), slog.Any("error", err))
	}
	return websocket.Message.Send(pc.conn, dataBytes)
}

func read(ws *websocket.Conn, logger *slog.Logger) {
	var msg = make([]byte, 512)
	var n int
	var err error
	if n, err = ws.Read(msg); err != nil {
		logger.Error("Could not read data", "error", err)
	}
	fmt.Printf("Message from host: %s.\n", msg[:n])
}
