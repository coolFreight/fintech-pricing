package pricing

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/gorilla/mux"
	gSocket "github.com/gorilla/websocket"
)

var upgrader = gSocket.Upgrader{}

type PriceSimulator struct {
	ticker          string
	quotes          []EquityQuote
	randomizeQuotes bool
	pricingClient   *PricingClient
	ws              *gSocket.Conn
	tradeWs         *gSocket.Conn
	WebsockUrl      string
	HttpUrl         string
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

func NewPriceSimulator() (*PriceSimulator, error) {
	simulator := &PriceSimulator{}
	websockUrl, httpUrl, err := simulator.Start()
	if err != nil {
		slog.Error("could not start price simulator", slog.Any("error", err))
		return nil, err
	}
	simulator.WebsockUrl = websockUrl
	simulator.HttpUrl = httpUrl

	return simulator, err
}

func (ps *PriceSimulator) PublishPrice(quote EquityQuote) error {
	quotes := make([]EquityQuote, 0)
	quotes = append(quotes, quote)
	err := ps.ws.WriteJSON(quotes)
	if err != nil {
		slog.Error("Write JSON failed", slog.Any("error", err))
		return err
	} else {
		slog.Info("Publishing simulator price", slog.Any("price", quotes))
	}
	return nil
}

func (ps *PriceSimulator) PublishOrderEvent(trade any) error {
	err := ps.tradeWs.WriteJSON(trade)
	if err != nil {
		slog.Error("Write JSON failed", slog.Any("error", err))
		return err
	} else {
		slog.Info("Publishing simulated order event", slog.Any("order", trade))
	}
	return nil
}

func RandomizeStart(ticker string) (string, string, error) {
	simulator := &PriceSimulator{
		ticker:          ticker,
		randomizeQuotes: true,
	}
	return simulator.Start()
}

func (ps *PriceSimulator) Start() (string, string, error) {

	go func() {
		r := mux.NewRouter().StrictSlash(true)
		r.HandleFunc("/pricing", ps.priceSimulation)
		r.HandleFunc("/order-manager", ps.orderSimulation)
		r.HandleFunc("/v2/stocks/snapshots", func(writer http.ResponseWriter, request *http.Request) {
			slog.Info("Received snapshot request")
			//writer.WriteHeader(http.StatusOK)
			//writer.Header().Add("Content-Type", "application/json")
			snapShot := marketdata.Snapshot{
				LatestQuote: &marketdata.Quote{
					BidPrice: 100.00,
				},
				MinuteBar: &marketdata.Bar{
					Open:  100.00,
					High:  100.00,
					Low:   100.00,
					Close: 100.00,
				},
			}

			snapshots := make(map[string]*marketdata.Snapshot)
			snapshots["FACEPACA"] = &snapShot
			writer.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(writer).Encode(snapshots)
			if err != nil {
				slog.Error("could not marshal data ", slog.Any("data", snapShot), slog.Any("error", err))
			}
		})
		http.ListenAndServe(":8080", r)
		slog.Info("Pricing simulator websocket connected on", "websocket", "", "http", "s.URL")
	}()
	return "ws://127.0.0.1:8080/pricing", "http://127.0.0.1:8080", nil
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
		c.ReadJSON(&request)
		slog.Info("Pricing-Simulator: Received PricingConnect ", slog.Any("request", request))
		resp := &AuthResponse{
			Message: "authenticated",
			Status:  "success",
		}
		resps := make([]AuthResponse, 0)
		resps = append(resps, *resp)
		err = c.WriteJSON(resps)
	}
}

func (ps *PriceSimulator) orderSimulation(w http.ResponseWriter, r *http.Request) {
	slog.Info("Incoming order manager simulator request")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Websocket dial failed ", "error", err)
		return
	}

	ps.tradeWs = c
	for {
		var request Auth
		err = c.ReadJSON(&request)
		if err != nil {
			slog.Error("Could not read json data", "error", err)
		} else {
			slog.Info("Received Simulated Order Manager authentication request", "request", request)
			resp := &AuthResponse{
				Message: "authenticated",
				Status:  "success",
			}
			resps := make([]AuthResponse, 0)
			resps = append(resps, *resp)
			err = c.WriteJSON(resps)
			if err != nil {
				slog.Error("Could not send order manager simulator authenticated response", "error", err)
			}
		}
	}
}
