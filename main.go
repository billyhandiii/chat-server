package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	startChatServer(context.Background(), os.Args, os.Stdout)
}

func startChatServer(ctx context.Context, args []string, defaultLog messageReceiver) {
	log.Print("starting up")
	ctx = withSignalCancel(ctx)

	var (
		wg  sync.WaitGroup
		err error
		cfg *config
	)

	if len(args) > 1 {
		configFilename := args[1]
		cfg, err = readConfigFile(configFilename)
		if err != nil {
			log.Fatalf("could not read config file (%v): %v", configFilename, err)
		}
	} else {
		cfg = defaultConfig()
	}

	mb := newMessageBroker()

	wg.Add(1)
	go mb.eventLoop(&wg)

	if len(cfg.LogFilePath) > 0 {
		var logFile *os.File
		logFile, err = os.OpenFile(cfg.LogFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Fatalf("could not open log file for writing (%v): %v", cfg.LogFilePath, err)
		}

		mb.addReceiver(logFile)
	} else {
		mb.addReceiver(defaultLog)
	}

	wg.Add(1)
	go acceptConnections(ctx, &wg, cfg, mb)

	wg.Add(1)
	go acceptHTTPMessages(ctx, &wg, cfg, mb)

	<-ctx.Done()

	log.Print("shutting down")
	mb.close()

	wg.Wait()
}

func withSignalCancel(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()

	return ctx
}
