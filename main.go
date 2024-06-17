package main

import (
	"context"
	"fmt"
	"gossip-protocol/constants"
	"gossip-protocol/db"
	"gossip-protocol/network"
	"gossip-protocol/p2p"
	"gossip-protocol/processor"
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
	dbConfig := cfg.ReadDBConfig()
	dbURL := dbConfig.AsPostgresDbUrl()
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
	db := db.Init(dbURL)
	newTracker := network.NewMessageTracker(constants.MAX_MESSAGES, db)

	p2pNode.Start(&wg)
	wg.Add(1)
	go processor.Processor(ctx, &wg, obsReceiveRes, newTracker)
	wg.Wait()

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
