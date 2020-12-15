package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/elciok/swarmonitor/status"
)

func UpdateStatusList(statusList *status.StatusList) error {
	for _, status := range statusList.List {
		err := UpdateStatus(status)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateStatus(status *status.Status) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	filters := filters.NewArgs()
	for label, value := range status.Labels {
		labelString := fmt.Sprintf("%s=%s", label, value)
		filters.Add("label", labelString)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters,
	})
	if err != nil {
		return err
	}

	status.Running = false
	status.Healthy = false
	for _, container := range containers {
		if container.State == "running" {
			status.Running = true
		}

		if strings.Contains(container.Status, "(healthy)") {
			status.Healthy = true
		}

		if status.Ok() {
			break
		}
	}

	return nil
}
