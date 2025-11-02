package internal

import (
	"encoding/json"
	"log/slog"
	"os"

	"golang.org/x/net/websocket"
)

const (
	APCA_KEY                   = "APCA_API_KEY_ID"
	APCA_SECRET                = "APCA_API_SECRET_KEY"
	APCA_BASE_URL              = "APCA_BASE_URL"
	APCA_API_VERSION           = "APCA_API_VERSION"
	APCA_PRICING_MARKET_STREAM = "APCA_PRICING_MARKET_STREAM"
)

type Auth struct {
	Action string `json:"action"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func Connect() (*websocket.Conn, error) {
	origin := os.Getenv(APCA_BASE_URL) + os.Getenv(APCA_API_VERSION)
	url := os.Getenv(APCA_PRICING_MARKET_STREAM)
	logger.Info("configs ", slog.String(APCA_BASE_URL, os.Getenv(APCA_BASE_URL)))
	logger.Info("configs ", slog.String(APCA_PRICING_MARKET_STREAM, os.Getenv(APCA_PRICING_MARKET_STREAM)))
	logger.Info("configs ", slog.String(APCA_API_VERSION, os.Getenv(APCA_API_VERSION)))
	logger.Info("Connecting websocket stream ", slog.String("url", url))
	logger.Info("Using ", slog.String("origin", origin))

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		logger.Error("Could not connect to websocket", slog.Any("error", err))
		return nil, err
	}
	logger.Info("Successfully connected to pricing host")
	authenticate := Auth{
		Action: "auth",
		Key:    os.Getenv(APCA_KEY),
		Secret: os.Getenv(APCA_SECRET),
	}
	err = Send(authenticate, ws)
	if err != nil {
		logger.Error("could not authenticate - ", slog.Any("error", err))
		return nil, err
	}
	logger.Info("Successfully authenticated to pricing host")
	return ws, nil
}

func Send(data any, ws *websocket.Conn) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logger.Error("could not marshal data ", slog.Any("data", data), slog.Any("error", err))
	}
	return websocket.Message.Send(ws, dataBytes)
}
