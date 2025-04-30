package pricing

import (
	"encoding/json"
	"fmt"
	"github.com/coolFreight/fintech-pricing/internal"
	"golang.org/x/net/websocket"
	"log/slog"
	"time"
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

func NewQuotes(tickers []string, logger *slog.Logger) <-chan []EquityQuote {
	conn, err := internal.Connect()
	if err != nil {
		logger.Error("Error connect pricing", "error", err)
		return nil
	}
	logger.Info(fmt.Sprintf("Subscribing for pricing for tickers : [ %s ]", tickers))
	quotes := PricingConnect{Action: "subscribe", Quotes: tickers}
	err = send(quotes, conn, logger)
	if err != nil {
		logger.Error("Could not subscribe to quotes ", "error", err)
		return nil
	}
	read(conn, logger)
	logger.Info(fmt.Sprintf("Successfully subscribed for quotes %s", tickers))
	quoteChan := make(chan []EquityQuote)
	go func() {
		defer close(quoteChan)
		defer conn.Close()
		for {
			var quotes []EquityQuote
			if err := websocket.JSON.Receive(conn, &quotes); err != nil {
				if err != nil {
					logger.Error("Error receiving an update", "error", err)
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
			quoteChan <- quotes
		}
	}()
	return quoteChan
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

func NewCryptoPricing(conn *websocket.Conn, logger *slog.Logger) <-chan []CryptoQuote {
	quoteChan := make(chan []CryptoQuote)
	go func() {
		for {
			var quotes []CryptoQuote
			if err := websocket.JSON.Receive(conn, &quotes); err != nil {
				logger.Error("Could not get pricing ", "error", err)
				close(quoteChan)
				return
			}
			quoteChan <- quotes
		}
	}()
	return quoteChan
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

func read(ws *websocket.Conn, logger *slog.Logger) {
	var msg = make([]byte, 512)
	var n int
	var err error
	if n, err = ws.Read(msg); err != nil {
		logger.Error("Could not read data", "error", err)
	}
	fmt.Printf("Mesage from host: %s.\n", msg[:n])
}
