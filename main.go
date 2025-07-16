package main

import (
	"fintech-pricing/internal"
	"fintech-pricing/pricing"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
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
		_ = os.Setenv(internal.APCA_MARKET_PRICING_STREAM, os.Getenv("APCA_TEST_PRICING_MARKET_STREAM"))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Lets get some pricing")
	var wg sync.WaitGroup
	done := make(chan any)
	pc := pricing.NewPricingClient(done, []string{"FAKEPACA"}, logger)

	logger.Info("Lets get some pricing again")
	time.Sleep(time.Second * 30)
	pc.Subscribe([]string{"APPL"})

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Minute * 60)
		close(done)
		time.Sleep(time.Second * 30)
	}()

	wg.Wait()
}
