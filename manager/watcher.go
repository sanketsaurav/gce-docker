package manager

import (
	"fmt"
	"time"

	"github.com/fsouza/go-dockerclient"
)

var IPAssignLabel = "gce.driver.ip.static"

type Watcher struct {
	WatchedStatus map[string]bool
	WatchedLabels map[string]bool
	DefaultDelay  time.Duration

	c        *docker.Client
	w        *Worker
	listener chan *docker.APIEvents
}

func NewWatcher() (*Watcher, error) {
	c, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		WatchedStatus: map[string]bool{"die": true, "start": true},
		WatchedLabels: map[string]bool{IPAssignLabel: true},
		DefaultDelay:  time.Second * 1,
		c:             c,
		w:             NewWorker(),
	}, nil
}

func (m *Watcher) Start() error {
	m.listener = make(chan *docker.APIEvents, 0)

	if err := m.c.AddEventListener(m.listener); err != nil {
		return err
	}

	for e := range m.listener {
		m.handleEvent(e)
	}

	return nil
}

func (m *Watcher) handleEvent(e *docker.APIEvents) error {
	if !m.WatchedStatus[e.Status] {
		return nil
	}

	c, err := m.c.InspectContainer(e.ID)
	if err != nil {
		return err
	}

	labels := m.watchedLabels(c)
	if len(labels) == 0 {
		return nil
	}

	fmt.Println(e, c.Config.Labels, labels)

	switch e.Status {
	case "die":
		return m.detach(c, labels)
	case "start":
		return m.attach(c, labels)
	}

	return nil
}

func (m *Watcher) watchedLabels(c *docker.Container) map[string]string {
	var matched = make(map[string]string, 0)
	for label, value := range c.Config.Labels {
		if !m.WatchedLabels[label] {
			continue
		}

		matched[label] = value
	}

	return matched
}

func (m *Watcher) attach(c *docker.Container, l map[string]string) error {
	jobID := JobID(c.ID)

	m.w.Delete(jobID)
	m.w.Add(jobID, func() error {
		fmt.Println("attach", c.ID)
		return nil
	}, m.DefaultDelay)

	return nil
}

func (m *Watcher) detach(c *docker.Container, l map[string]string) error {
	jobID := JobID(c.ID)

	m.w.Delete(jobID)
	m.w.Add(JobID(c.ID), func() error {
		fmt.Println("dettach", c.ID)

		return nil
	}, m.DefaultDelay)

	return nil
}
