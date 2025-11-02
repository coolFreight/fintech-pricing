package pricing

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"

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
	conn        *websocket.Conn
	logger      *slog.Logger
	done        chan any
	Quotes      []string
	retryChan   chan bool
	pricingChan chan []EquityQuote
}

func NewPricingClient(conn *websocket.Conn, retryChan chan bool, logger *slog.Logger) *PricingClient {
	return &PricingClient{
		conn:        conn,
		logger:      logger,
		retryChan:   retryChan,
		pricingChan: make(chan []EquityQuote),
	}
}

func (pc *PricingClient) Start(tickers []string) (<-chan []EquityQuote, error) {
	logger := pc.logger
	logger.Info(fmt.Sprintf("Subscribing for pricing for tickers : %s ", tickers))
	err := send(PricingConnect{Action: "subscribe", Quotes: tickers}, pc.conn, logger)
	pc.Quotes = tickers
	if err != nil {
		logger.Error("Could not subscribe to quotes ", "error", err)
		return nil, errors.New("Could not subscribe to quotes ")
	}
	read(pc.conn, logger)
	logger.Info(fmt.Sprintf("Successfully subscribed for quotes %s", tickers))
	pc.pricing()
	return pc.pricingChan, nil
}

func (pc *PricingClient) pricing() {
	var quotes []EquityQuote
	go func() {
		defer logger.Warn("shutting down pricing websocket connection")
		defer func(conn *websocket.Conn) {
			err := conn.Close()
			if err != nil {
				logger.Error("Error pricing closing connection", "error", err)
			}
		}(pc.conn)
		logger.Info("####### Starting pricing quotes #######")
		for {
			if err := websocket.JSON.Receive(pc.conn, &quotes); err != nil {
				if err == io.EOF {
					logger.Warn("Pricing connection lost")
					pc.retryChan <- true
					return
				} else {
					logger.Error("Error reading data", "error", err)
					return
				}
			} else {
				pc.pricingChan <- quotes
			}
		}
	}()
}

func (pc *PricingClient) Reconnect() (*websocket.Conn, error) {
	pc.logger.Info("Reconnecting for pricing")
	conn, err := internal.Connect()
	if err != nil {
		logger.Error("Could not connect for pricing", "error", err)
	} else {
		quotes := PricingConnect{Action: "subscribe", Quotes: pc.Quotes}
		err = send(quotes, conn, logger)
		if err != nil {
			logger.Error("Could not subscribe to quotes ", "error", err)
		} else {
			read(conn, logger)
			pc.conn = conn
			pc.pricing()
			logger.Info("Successfully resubscribed for price quotes")
		}
	}
	return pc.conn, err
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
		return err
	}
	return websocket.Message.Send(ws, dataBytes)
}

func (pc *PricingClient) UpdateSubscription(ticker string) error {
	pc.Quotes = append(pc.Quotes, ticker)
	pc.logger.Info("Subscribing for pricing for tickers", slog.Any("tickers", ticker))
	dataBytes, err := json.Marshal(PricingConnect{Action: "subscribe", Quotes: pc.Quotes})
	if err != nil {
		pc.logger.Error("could subscribe for tickers", slog.Any("tickers", pc.Quotes), slog.Any("error", err))
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

func (pc *PricingClient) Stop() {
	slog.Info("Stopping Price Simulator")
	defer close(pc.pricingChan)
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(pc.conn)
	pc.done <- "nil"
}
