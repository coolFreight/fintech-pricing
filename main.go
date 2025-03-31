package main

import (
	"encoding/json"
	"fmt"
	"github.com/coolFreight/fintech-pricing/pricing"
	"golang.org/x/net/websocket"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

var prefix = "APCA_PAPER"

type auth struct {
	Action string `json:"action"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if os.Getenv("APCA_ENVIRONMENT") == "prod" {
		logger.Info("******* LIVE ENVIRONMENT ****************")
		prefix = "APCA_LIVE"
	} else {
		logger.Info("******* PAPER ENVIRONMENT ****************")
	}

	logger.Info(fmt.Sprintf("calling alpaca %s", os.Getenv(prefix+"_BASE_URL")))

	_ = os.Setenv(strings.Split(prefix, "_")[0]+"_API_KEY_ID", os.Getenv(prefix+"_API_KEY"))
	_ = os.Setenv(strings.Split(prefix, "_")[0]+"_API_SECRET_KEY", os.Getenv(prefix+"API_SECRET"))
	_ = os.Setenv(strings.Split(prefix, "_")[0]+"_BASE_URL", os.Getenv(prefix+"_BASE_URL"))
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Starting to connect for pricing")
	ws, err := startStream(prefix, logger)

	count := 5
	for count > 0 {

		logger.Info("Attempting to authenticate")
		authenticate := auth{
			Action: "auth",
			Key:    os.Getenv(prefix + "_API_KEY"),
			Secret: os.Getenv(prefix + "_API_SECRET"),
		}
		err = send(authenticate, ws, logger)
		if err != nil {
			log.Fatal("could not authenticate - ", err)
		}
		read(ws, logger)
		//
		//fmt.Printf("Requesting trades\n")
		//connect := pricing.TradeConnect{Action: "subscribe", Trades: []string{"IBM"}}
		//err = send(connect, ws, logger)
		//read(ws, logger)

		//quotes := pricing.PricingConnect{Action: "subscribe", Quotes: []string{"FAKEPACA"}}
		//err = send(quotes, ws, logger)
		//read(ws, logger)

		//var wg sync.WaitGroup
		//trades := pricing.NewTrades(ws, logger)
		//wg.Add(1)
		//go func() {
		//	defer wg.Done()
		//	for trade := range trades {
		//		fmt.Println(trade)
		//	}
		//}()

		var wg sync.WaitGroup
		quotesPricing := pricing.NewQuotes(ws, []string{"FAKEPACA"}, logger)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for quote := range quotesPricing {
				fmt.Println(quote)
			}
		}()
		wg.Wait()
		count--
		logger.Info(fmt.Sprintf("Stream closed retrying %d", count))
		time.Sleep(5 * time.Second)
	}

}

func startStream(prefix string, logger *slog.Logger) (*websocket.Conn, error) {
	logger.Info("Starting websocket stream .......")
	origin := os.Getenv(prefix+"_BASE_URL") + os.Getenv(prefix+"_API_VERSION")
	url := os.Getenv(prefix + "_MARKET_STREAM")

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	read(ws, logger)
	return ws, nil
}

func send(data any, ws *websocket.Conn, logger *slog.Logger) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logger.Error(fmt.Sprintf("could not marshal data %s", data, err))
	}
	return websocket.Message.Send(ws, dataBytes)
}

func read(ws *websocket.Conn, logger *slog.Logger) {
	var msg = make([]byte, 512)
	var n int
	var err error
	if n, err = ws.Read(msg); err != nil {
		logger.Error("Could not read data", err)
	}
	fmt.Printf("Received: %s.\n", msg[:n])
}
