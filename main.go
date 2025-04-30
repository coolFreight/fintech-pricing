package main

import (
	"flag"
	"fmt"
	"github.com/coolFreight/fintech-pricing/internal"
	"github.com/coolFreight/fintech-pricing/pricing"
	"log/slog"
	"os"
	"sync"
)

const (
	APCA_KEY                   = "APCA_API_KEY_ID"
	APCA_SECRET                = "APCA_API_SECRET_KEY"
	APCA_BASE_URL              = "APCA_BASE_URL"
	APCA_API_VERSION           = "APCA_API_VERSION"
	APCA_MARKET_STREAM         = "APCA_MARKET_STREAM"
	APCA_MARKET_HISTORICAL_URL = "APCA_MARKET_HISTORICAL_URL"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	envPtr := flag.String("env", "dev", "The environment to run in (test, dev, prod)")
	var prefix = "APCA_PAPER"
	flag.Parse()
	env := *envPtr
	if env == "prod" {
		logger.Info("******* LIVE ENVIRONMENT ****************")
		prefix = "APCA_LIVE"
	} else {
		logger.Info("******* PAPER ENVIRONMENT ****************")
	}

	logger.Info(fmt.Sprintf("Program args %s", env))
	_ = os.Setenv(APCA_KEY, os.Getenv(prefix+"_API_KEY"))
	_ = os.Setenv(APCA_SECRET, os.Getenv(prefix+"_API_SECRET"))
	_ = os.Setenv(APCA_BASE_URL, os.Getenv(prefix+"_BASE_URL"))
	_ = os.Setenv(APCA_API_VERSION, os.Getenv(prefix+"_API_VERSION"))
	_ = os.Setenv(internal.APCA_MARKET_PRICING_STREAM, os.Getenv(prefix+"_PRICING_MARKET_STREAM"))
	_ = os.Setenv(APCA_MARKET_HISTORICAL_URL, os.Getenv(prefix+"_MARKET_HISTORICAL_URL"))
	if env == "test" {
		logger.Info("******* PRICING FOR FAKEPACA TEST ENVIRONMENT ****************")
		_ = os.Setenv(APCA_MARKET_STREAM, os.Getenv("APCA_TEST_PRICING_MARKET_STREAM"))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Lets get some pricing")
	var wg sync.WaitGroup
	quoteChan := pricing.NewQuotes([]string{"SPXS", "IBM", "YOU", "FAKEPACA"}, logger)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for quote := range quoteChan {
			fmt.Println(quote)
		}
	}()
	wg.Wait()
}
