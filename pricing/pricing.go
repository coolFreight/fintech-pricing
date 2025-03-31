package pricing

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log/slog"
)

type EquityTrade struct {
	MessageType  string   `json:"T"`
	Symbol       string   `json:"S"`
	TradeId      int      `json:"i"`
	ExchangeCode string   `json:"x"`
	Price        float64  `json:"p"`
	Size         int      `json:"s"`
	Condition    []string `json:"c"`
	Timestamp    string   `json:"t"`
	Tape         string   `json:"z"`
}

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

func NewTrades(conn *websocket.Conn, logger *slog.Logger) <-chan []EquityTrade {
	quoteChan := make(chan []EquityTrade)
	go func() {
		for {
			var quotes []EquityTrade
			if err := websocket.JSON.Receive(conn, &quotes); err != nil {
				logger.Error("Could not get trades ", err)
				close(quoteChan)
				return
			}
			quoteChan <- quotes
		}
	}()
	return quoteChan
}

func NewQuotes(conn *websocket.Conn, logger *slog.Logger) <-chan []EquityQuote {
	quoteChan := make(chan []EquityQuote)
	go func() {
		for {
			var quotes []EquityQuote
			if err := websocket.JSON.Receive(conn, &quotes); err != nil {
				logger.Error("Could not get pricing ", err)
				close(quoteChan)
				return
			}
			quoteChan <- quotes
		}
	}()
	return quoteChan
}

func NewCryptoPricing(conn *websocket.Conn, logger *slog.Logger) <-chan []CryptoQuote {
	quoteChan := make(chan []CryptoQuote)
	go func() {
		for {
			var quotes []CryptoQuote
			if err := websocket.JSON.Receive(conn, &quotes); err != nil {
				logger.Error("Could not get pricing ", err)
				close(quoteChan)
				return
			}
			quoteChan <- quotes
		}
	}()
	return quoteChan
}

func (t EquityTrade) String() string {
	return fmt.Sprintf("EquityTrade"+
		"{Ticker: %s, "+
		"Price: %f, Size: %d,Time: %s, TradeId: %d}", t.Symbol, t.Price, t.Size, t.Timestamp, t.TradeId)
}

func (q EquityQuote) String() string {
	return fmt.Sprintf("EquityQuote"+
		"{Ticker: %s, "+
		"BidPrice: %f, BidSize: %d, BidExchange: %s, AskPrice: %f, AskSize: %d, AskEchange: %s, Time: %s}",
		q.Symbol, q.BidPrice, q.BidSize, q.BidExchange, q.AskPrice, q.AskSize, q.AskExchange, q.Timestamp)
}
