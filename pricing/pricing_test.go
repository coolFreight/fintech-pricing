package pricing

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/coolFreight/fintech-pricing/internal"
	"github.com/datenhahn/golang-awaitility/awaitility"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func TestPricingReconnect(t *testing.T) {
	simulator := NewPriceSimulator()
	err := simulator.Start()
	if err != nil {
		t.Errorf("Failing test")
	}

	retryChan := make(chan bool)

	url := simulator.WebsockUrl

	_ = os.Setenv(internal.APCA_BASE_URL, url)
	_ = os.Setenv(internal.APCA_PRICING_MARKET_STREAM, url)
	time.Sleep(1 * time.Second)
	pricingWebsocket, err := internal.Connect()
	pricingClient := NewPricingClient(pricingWebsocket, retryChan, logger)
	_, err = pricingClient.Start([]string{"ACA"})
	if err != nil {
		t.Errorf("Failing test")
	}

	go func() {
		for {
			select {
			case <-retryChan:
				_, err = pricingClient.Reconnect()
				if err != nil {
					t.Errorf("Failing test")
				}
				logger.Info("Called reconnect on pricing client.")
			}
		}
	}()

	if err != nil {
		slog.Error("Error connecting to pricing", slog.Any("error", err))
		return
	}

	//give some time to do the subcription subcribes
	time.Sleep(2 * time.Second)
	go func() {
		for prices := range pricingClient.pricingChan {
			for _, price := range prices {
				slog.Info("Received price", slog.Any("quote", price))
			}
		}
	}()

	err = simulator.PublishPrice([]EquityQuote{{BidPrice: 75.46, AskPrice: 65.00, Symbol: "ACA"}})
	if err != nil {
		t.Errorf("Failing test")
	}
	time.Sleep(500 * time.Millisecond)
	err = simulator.PublishPrice([]EquityQuote{{BidPrice: 85.46, AskPrice: 95.00, Symbol: "ACA"}})
	if err != nil {
		t.Errorf("Failing test")
	}
	time.Sleep(500 * time.Millisecond)
	err = simulator.PublishPrice([]EquityQuote{{BidPrice: 985.46, AskPrice: 195.00, Symbol: "ACA"}})
	if err != nil {
		t.Errorf("Failing test")
	}

	simulator.ServerConnectionClose()

	err = awaitility.Await(1000*time.Millisecond, 5*time.Second, func() bool {
		err = simulator.PublishPrice([]EquityQuote{{BidPrice: 1175.46, AskPrice: 6435.00, Symbol: "ACA"}})
		return err == nil
	})

	if err != nil {
		t.Errorf("Expect a reconnection to happen")
	}

}
