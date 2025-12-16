package pricing

import (
	"log/slog"

	"golang.org/x/net/websocket"
)

type MockClient struct{}

func (m MockClient) Start(_ []string) (<-chan []EquityQuote, error) {
	pc := make(chan []EquityQuote)
	return pc, nil
}

func (m MockClient) Reconnect() (*websocket.Conn, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockClient) UpdateSubscription(ticker string) error {
	logger.Info("mock update subscription", slog.String("ticker", ticker))
	return nil
}

func (m MockClient) Stop() {
	//TODO implement me
	panic("implement me")
}
