package main

import (
	"errors"
	"fmt"

	"github.com/elciok/swarmonitor/config"
	"github.com/elciok/swarmonitor/docker"
	"github.com/elciok/swarmonitor/notifier"
	"github.com/elciok/swarmonitor/status"
)

func main() {
	config := config.ReadConfig()

	statusList := status.NewStatusList(config.ContainerDir)
	err := statusList.ReadFromFiles()
	if err != nil {
		panic(err)
	}
	err = docker.UpdateStatusList(statusList)
	if err != nil {
		panic(err)
	}

	allRunning := true
	for _, containerStatus := range statusList.List {
		if !containerStatus.Ok() {
			err = notifier.SendNotification(config, containerStatus)
			if err != nil {
				panic(err)
			}
			allRunning = false
			fmt.Println("No running containers with labels:", containerStatus.Labels)
		}
	}
	if !allRunning {
		panic(errors.New("No running containers for at least one of the container sets checked."))
	}
	fmt.Println("swarmonitor - checks completed.")
}
