package main

import (
	"context"
	"fmt"
	"gossip-protocol/network"
	"gossip-protocol/p2p"
	"os"
	"os/signal"
	"sync"
	"time"

	"gossip-protocol/config"

	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.NewViperConfig()
	var wg sync.WaitGroup // Create a WaitGroup

	obsReceiveRes := make(chan network.Message, 1)
	obsSendReq := make(chan network.Message, 1)
	loggerLevel := cfg.ReadLogLevelConfig()
	logger := NewRootLogger(loggerLevel, cfg)

	p2pCfg := cfg.ReadP2PConfig()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sign := <-c
		logger.Infof("System signal: %+v\n", sign)
		cancel()
	}()

	p2pConfig := p2p.ToP2pConfig(p2pCfg)

	p2pNode := p2p.NewP2PNode(ctx, logger.WithField("layer", "p2p"), obsReceiveRes, obsSendReq, p2pConfig)

	p2pNode.Start(&wg)

}

func NewRootLogger(loggerLevel string, cfg config.Config) *logrus.Logger {
	// Initialize root logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     false,
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	})

	logLevel, err := logrus.ParseLevel(loggerLevel)
	if err != nil {
		panic(fmt.Errorf("failed to parse log level: %v", err))
	}
	logger.SetLevel(logLevel)

	return logger
}

type NodeDetails struct {
	WorkerAddress    string   `json:"worker_address"`
	PrivateKey       string   `json:"private_key"`
	SingleMarketKeys []string `json:"single_market_keys"`
}

func ToNodeDetailsConfig(config config.NodeDetailsConfig) NodeDetails {
	return NodeDetails{
		WorkerAddress:    config.WorkerAddress,
		PrivateKey:       config.PrivateKey,
		SingleMarketKeys: config.SingleMarketKeys,
	}
}
