package pricing

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/gorilla/mux"
	gSocket "github.com/gorilla/websocket"
	"golang.org/x/net/context"
)

var upgrader = gSocket.Upgrader{}
var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

type PriceSimulator struct {
	ticker          string
	quotes          []EquityQuote
	randomizeQuotes bool
	pricingClient   *PricingClient
	ws              *gSocket.Conn
	tradeWs         *gSocket.Conn
	Snapshots       []marketdata.Snapshot
	WebsockUrl      string
	HttpUrl         string
	endpoints       map[string]func(writer http.ResponseWriter, request *http.Request)
	server          *http.Server
}

type Auth struct {
	Action string `json:"action"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type OrderEvent struct {
	Event     string
	EventId   string
	Order     alpaca.Order `gorm:"type:jsonb" json:"order"`
	At        time.Time
	Timestamp time.Time
}

type TradeUpdate struct {
	Stream       string       `json:"stream"`
	ResponseData ResponseData `json:"data"`
}

type ResponseData struct {
	At        string       `json:"at"`
	EventId   string       `json:"event_id"`
	Event     string       `json:"event"`
	Timestamp string       `json:"timestamp"`
	Order     alpaca.Order `json:"order"`
}

type AuthResponse struct {
	Message string `json:"message"`
	Status  string `json:"T"`
}

func NewPriceSimulator() *PriceSimulator {
	simulator := &PriceSimulator{
		endpoints: make(map[string]func(writer http.ResponseWriter, request *http.Request)),
	}
	return simulator
}

func (ps *PriceSimulator) PublishPrice(quote any) error {

	if quote == io.EOF {
		logger.Info("received io.EOF for pricing")
		err := ps.ws.WriteMessage(gSocket.CloseMessage, gSocket.FormatCloseMessage(gSocket.CloseNormalClosure, ""))
		if err != nil {
			logger.Error("Close pricing  ws failed", slog.Any("error", err))
			return err
		}
	} else {
		q := quote.(EquityQuote)
		quotes := make([]EquityQuote, 0)
		quotes = append(quotes, q)
		err := ps.ws.WriteJSON(quotes)
		if err != nil {
			logger.Error("Write JSON failed", slog.Any("error", err))
			return err
		} else {
			logger.Info("Publishing simulator price", slog.Any("price", quotes))
		}

	}
	return nil
}

func (ps *PriceSimulator) PublishOrderEvent(trade any) error {

	if trade == io.EOF {
		logger.Info("received io.EOF")
		err := ps.tradeWs.WriteMessage(gSocket.CloseMessage, gSocket.FormatCloseMessage(gSocket.CloseNormalClosure, ""))
		if err != nil {
			logger.Error("Close trade ws failed", slog.Any("error", err))
			return err
		}
	} else {
		logger.Info("Creating a trade Json")
		err := ps.tradeWs.WriteJSON(trade)
		if err != nil {
			logger.Error("Write JSON failed", slog.Any("error", err))
			return err
		} else {
			logger.Info("Publishing simulated order event", slog.Any("order", trade))
		}
	}
	return nil
}

func (ps *PriceSimulator) Register(endpoint string, f func(writer http.ResponseWriter, request *http.Request)) {
	ps.endpoints[endpoint] = f
}

func (ps *PriceSimulator) Start() error {
	logger.Info("Starting PriceSimulator")
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/pricing", ps.priceSimulation)
	r.HandleFunc("/order-manager", ps.orderSimulation)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		for endpoint, handler := range ps.endpoints {
			r.HandleFunc(endpoint, handler)
		}
		err := s.ListenAndServe()
		logger.Error("Server has shutdown", slog.Any("error", err))
	}()

	s.RegisterOnShutdown(func() {

	})
	ps.server = s
	ps.WebsockUrl = "ws://127.0.0.1:8080/pricing"
	ps.HttpUrl = "http://127.0.0.1:8080"
	slog.Info("Pricing simulator websocket connected on", "websocket", ps.WebsockUrl, "http", ps.HttpUrl)
	return nil
}

func (ps *PriceSimulator) Stop() {
	logger.Info("Stopping pricing simulator graceful with interrupt")
	err := ps.server.Shutdown(context.Background())

	if err != nil {
		slog.Error("Failed to close http server", slog.Any("error", err))
		return
	}
}

func (ps *PriceSimulator) priceSimulation(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Websocket dial failed ", "error", err)
		return
	}

	ps.ws = c
	for {
		var request PricingConnect
		err = c.ReadJSON(&request)
		if err != nil {
			slog.Info("Pricing Websocket read error", "error", err)
			return
		}

		logger.Info("Pricing-Simulator: Received PricingConnect ", slog.Any("request", request))
		resp := &AuthResponse{
			Message: "authenticated",
			Status:  "success",
		}
		resps := make([]AuthResponse, 0)
		resps = append(resps, *resp)
		err = c.WriteJSON(resps)
		if err != nil {
			logger.Error("Pricing-Simulator: Websocket write error", slog.Any("error", err))
		}
		time.Sleep(3 * time.Second)
	}
}

func (ps *PriceSimulator) orderSimulation(w http.ResponseWriter, r *http.Request) {
	logger.Info("Incoming order manager simulator request")
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		slog.Error("Websocket dial failed ", "error", err)
		return
	}

	ps.tradeWs = c
	for {

		_, message, err := c.ReadMessage()
		if err != nil {
			slog.Error("Could not read json data", "error", err)
		} else {
			slog.Info("Received Simulated Order Manager authentication request", "request", message)
			err = c.WriteMessage(gSocket.TextMessage, message)
			if err != nil {
				slog.Error("Could not send order manager simulator authenticated response", "error", err)
			}
		}
	}
}
