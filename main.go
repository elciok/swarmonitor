package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elciok/swarmonitor/config"
	"github.com/elciok/swarmonitor/docker"
	"github.com/elciok/swarmonitor/notifier"
	"github.com/elciok/swarmonitor/status"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer func() {
		// signal.Stop(signalChan)
		cancel()
	}()

	log.SetOutput(os.Stdout)

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg := config.ReadConfig()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(cfg.TickInterval):
			cfg = config.ReadConfig()

			statusList := status.NewStatusList(cfg.ContainerDir)

			if err := statusList.ReadFromFiles(); err != nil {
				return err
			}
			if err := docker.UpdateStatusList(statusList); err != nil {
				return err
			}

			for _, containerStatus := range statusList.List {
				if !containerStatus.Ok() {
					log.Printf("No running containers in %s.", containerStatus.Target)
					if err := notifier.SendNotification(cfg, containerStatus); err != nil {
						return err
					}
				}
			}

			log.Println("swarmonitor - checks completed.")
		}
	}
}
