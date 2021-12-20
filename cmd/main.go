package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
	"xtest/server"
)

func main() {
	log.Println("Daemon started")
	cfgPath, err := server.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := server.NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	func() {
		defer func() {
			if e := recover(); e != nil {
				log.Fatalf("Recovered with panic: %v\nStack trace:\n%s\n", e, debug.Stack())
			}
		}()
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		s := server.NewServer(cfg)
		if err := s.Start(); err != nil {
			log.Fatal(err)
		}

	mainLoop:
		for {
			select {
			case sig := <-ch:
				log.Printf("OS Signal received: %v", sig)
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer func() {
					cancel()
				}()
				if err := s.Srv.Shutdown(ctx); err != nil {
					log.Fatalf("Server Shutdown Failed: %+v", err)
				}
				break mainLoop
			}
		}
		s.Close()
		log.Println("Server shutdown")
	}()

	log.Println("Daemon stopped")
}
