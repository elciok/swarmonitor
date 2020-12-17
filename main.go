package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elciok/swarmonitor/config"
	"github.com/elciok/swarmonitor/docker"
	"github.com/elciok/swarmonitor/status"
)

const VERSION = "0.0.2"

func main() {
	showVersion := flag.Bool("version", false, "show application version")
	flag.Parse()
	if *showVersion {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer func() {
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
	statusList := status.NewStatusList(cfg.ContainerDir)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(cfg.TickInterval):
			cfg = config.ReadConfig()

			statusList.DataDir = cfg.ContainerDir
			statusList.Origin = cfg.Origin
			if err := statusList.ReadFromFiles(); err != nil {
				return err
			}
			if err := docker.UpdateStatusList(statusList); err != nil {
				return err
			}

			for _, containerStatus := range statusList.List {
				if containerStatus.ShouldSendNotification() {
					log.Printf("Sending notification. Containers in %s. Status: %s", containerStatus.Target, containerStatus.StatusString())
					if err := containerStatus.SendNotification(cfg.SMTP); err != nil {
						return err
					}
				}
			}

			log.Println("swarmonitor - checks completed.")
		}
	}
}
