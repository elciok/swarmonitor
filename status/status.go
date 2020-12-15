package status

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/elciok/swarmonitor/notifier"
	"gopkg.in/yaml.v2"
)

type Status struct {
	Target      string
	Labels      map[string]string
	CheckHealth bool
	Running     bool
	Healthy     bool
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
	return &Status{
		Target:  target,
		Labels:  labels,
		Running: false,
		Healthy: false,
	}
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
			status.CheckHealth = filenameHasCheckHealthFlag(file)

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

func (status *Status) Ok() bool {
	return (status.CheckHealth && status.Healthy) || (!status.CheckHealth && status.Running)
}

func (status *Status) SendNotification(cfg *notifier.SMTPConfig) error {
	subject := fmt.Sprintf("swarmonitor - %s is %s", status.Target, statusString(status))
	body := bodyString(status)
	if err := notifier.SendNotification(cfg, subject, body); err != nil {
		return err
	}
	return nil
}

func bodyString(status *Status) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Status: %s\r\n\r\n", statusString(status))
	fmt.Fprint(&builder, "Labels:\r\n")
	for labelKey, labelValue := range status.Labels {
		fmt.Fprintf(&builder, "\t- %s = %s\r\n", labelKey, labelValue)
	}
	return builder.String()
}

func statusString(status *Status) string {
	if status.Ok() {
		return "OK"
	} else {
		return "DOWN"
	}
}
