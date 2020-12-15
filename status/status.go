package status

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/elciok/swarmonitor/notifier"
	"gopkg.in/yaml.v2"
)

type Status struct {
	Target      string
	Labels      map[string]string
	checkHealth bool
	running     bool
	healthy     bool
	changed     bool
}

type StatusList struct {
	DataDir string
	List    map[string]*Status
}

func NewStatusList(dataDir string) *StatusList {
	return &StatusList{
		DataDir: dataDir,
		List:    map[string]*Status{},
	}
}

func NewStatus(target string, labels map[string]string) *Status {
	status := &Status{
		Target:      target,
		Labels:      labels,
		checkHealth: filenameHasCheckHealthFlag(target),
		running:     false,
		healthy:     false,
		changed:     false,
	}

	// set to true because it is the expected state
	// if containers are down it will send notification
	if status.checkHealth {
		status.healthy = true
	} else {
		status.running = true
	}

	return status
}

func (statusList *StatusList) ReadFromFiles() error {
	var files map[string]bool = make(map[string]bool)

	err := filepath.Walk(statusList.DataDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yml" || filepath.Ext(path) == ".yaml" {
			files[path] = true
		}
		return nil
	})
	if err != nil {
		return err
	}
	for file, _ := range files {
		yamlFile, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		var labels map[string]string
		err = yaml.Unmarshal(yamlFile, &labels)
		if err != nil {
			return err
		}

		status, ok := statusList.List[file]
		if !ok {
			status = NewStatus(file, labels)

			statusList.List[file] = status
		} else {
			status.Labels = labels
		}
	}

	//remove files that were removed
	for file, _ := range statusList.List {
		_, ok := files[file]
		if !ok {
			delete(statusList.List, file)
		}
	}

	return nil
}

func filenameHasCheckHealthFlag(filename string) bool {
	filenameParts := strings.Split(filename, ".")
	if len(filenameParts) >= 3 {
		penultimate := filenameParts[len(filenameParts)-2]
		penultimate = strings.ToLower(penultimate)
		return penultimate == "healthy" || penultimate == "health"
	} else {
		return false
	}
}

func (status *Status) SetRunning(running bool) {
	if running != status.running && !status.checkHealth {
		status.changed = true
	}
	status.running = running
}

func (status *Status) SetHealthy(healthy bool) {
	if healthy != status.healthy && status.checkHealth {
		status.changed = true
	}
	status.healthy = healthy
}

func (status *Status) Ok() bool {
	return status.OkWithValues(status.running, status.healthy)
}

func (status *Status) OkWithValues(running bool, healthy bool) bool {
	return (status.checkHealth && healthy) || (!status.checkHealth && running)
}

func (status *Status) ShouldSendNotification() bool {
	return status.changed
}

func (status *Status) SendNotification(cfg *notifier.SMTPConfig) error {
	subject := fmt.Sprintf("swarmonitor - %s is %s", status.Target, status.StatusString())
	body := bodyString(status)
	if err := notifier.SendNotification(cfg, subject, body); err != nil {
		return err
	}
	status.changed = false
	return nil
}

func bodyString(status *Status) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Status: %s\r\n", status.StatusString())
	fmt.Fprintf(&builder, "Time: %s\r\n\r\n", time.Now())
	fmt.Fprint(&builder, "Labels:\r\n")
	for labelKey, labelValue := range status.Labels {
		fmt.Fprintf(&builder, "\t- %s = %s\r\n", labelKey, labelValue)
	}
	return builder.String()
}

func (status *Status) StatusString() string {
	if status.Ok() {
		return "OK"
	} else {
		return "DOWN"
	}
}
