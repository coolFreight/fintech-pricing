package pricing

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/coolFreight/fintech-pricing/internal"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func TestPricingReconnect(t *testing.T) {
	simulator := NewPriceSimulator()
	err := simulator.Start()
	if err != nil {
		t.Fail()
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
		t.Fail()
	}

	go func() {
		for {
			select {
			case <-retryChan:
				_, err = pricingClient.Reconnect()
				if err != nil {
					t.Fail()
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

	err = simulator.PublishPrice(EquityQuote{BidPrice: 75.46, AskPrice: 65.00, Symbol: "ACA"})
	if err != nil {
		t.Fail()
	}
	time.Sleep(500 * time.Millisecond)
	err = simulator.PublishPrice(EquityQuote{BidPrice: 85.46, AskPrice: 95.00, Symbol: "ACA"})
	if err != nil {
		t.Fail()
	}
	time.Sleep(500 * time.Millisecond)
	err = simulator.PublishPrice(EquityQuote{BidPrice: 985.46, AskPrice: 195.00, Symbol: "ACA"})
	if err != nil {
		t.Fail()
	}

	simulator.ServerConnectionClose()

	err = simulator.PublishPrice(EquityQuote{BidPrice: 1175.46, AskPrice: 6435.00, Symbol: "ACA"})
	if err != nil {
		t.Fail()
	}
	time.Sleep(500 * time.Millisecond)
	err = simulator.PublishPrice(EquityQuote{BidPrice: 1285.46, AskPrice: 9555.00, Symbol: "ACA"})
	if err != nil {
		t.Fail()
	}
	time.Sleep(500 * time.Millisecond)
}
