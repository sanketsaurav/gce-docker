package watcher

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/fsouza/go-dockerclient"
	"github.com/mcuadros/docker-volume-gce/providers"
)

var (
	LabelNetworkPrefix          = "gce."
	LabelNetworkType            = LabelNetworkPrefix + "lb.type"
	LabelNetworkGroup           = LabelNetworkPrefix + "lb.group"
	LabelNetworkAddress         = LabelNetworkPrefix + "lb.address"
	LabelNetworkSourceRanges    = LabelNetworkPrefix + "lb.source.ranges"
	LabelNetworkSourceTags      = LabelNetworkPrefix + "lb.source.tags"
	LabelNetworkSessionAffinity = LabelNetworkPrefix + "lb.session.affinity"
)

var validLabels = []string{
	LabelNetworkType, LabelNetworkGroup, LabelNetworkAddress,
	LabelNetworkSourceRanges, LabelNetworkSourceTags, LabelNetworkSessionAffinity,
}

type Watcher struct {
	WatchedStatus       map[string]bool
	WatchedLabelsPrefix string
	DefaultDelay        time.Duration

	c        *docker.Client
	p        *providers.Network
	w        *Worker
	listener chan *docker.APIEvents
}

func NewWatcher(d *docker.Client, c *http.Client, project, zone, instance string) (*Watcher, error) {
	p, err := providers.NewNetwork(c, project, zone, instance)
	if err != nil {
		return nil, err
	}

	return &Watcher{
		WatchedStatus:       map[string]bool{"die": true, "start": true},
		WatchedLabelsPrefix: LabelNetworkPrefix,
		DefaultDelay:        time.Second * 1,
		c:                   d,
		p:                   p,
		w:                   NewWorker(),
	}, nil
}

func (m *Watcher) Watch() error {
	m.listener = make(chan *docker.APIEvents, 0)

	if err := m.c.AddEventListener(m.listener); err != nil {
		return err
	}

	for e := range m.listener {
		if err := m.handleEvent(e); err != nil {
			log15.Error("error handling event", "container", e.ID[:12], "error", err)
		}
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

	log15.Debug("event captured", "status", e.Status, "container", e.ID[:12], "labels", labels)

	if err := m.validateLabels(labels); err != nil {
		return err
	}

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
		if !strings.HasPrefix(label, m.WatchedLabelsPrefix) {
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
		start := time.Now()
		config := m.createNetworkConfig(c, l)
		log15.Debug("start event detected, creating network",
			"container", c.ID[:12], "ports", config.Ports,
		)

		if err := m.p.Create(config); err != nil {
			log15.Error("error creating network",
				"container", c.ID[:12], "ports", config.Ports, "error", err,
			)
			return nil
		}

		log15.Info(
			"network started",
			"container", c.ID[:12], "ports", config.Ports, "elapsed", time.Since(start),
		)
		return nil
	}, m.DefaultDelay)

	return nil
}

func (m *Watcher) detach(c *docker.Container, l map[string]string) error {
	jobID := JobID(c.ID)

	m.w.Delete(jobID)
	m.w.Add(JobID(c.ID), func() error {
		start := time.Now()
		config := m.createNetworkConfig(c, l)
		log15.Debug("stop event detected, deleting network",
			"container", c.ID[:12], "ports", config.Ports,
		)

		if err := m.p.Delete(config); err != nil {
			log15.Error("error deleting network",
				"container", c.ID[:12], "ports", config.Ports, "error", err,
			)

			return nil
		}

		log15.Info(
			"network deleted",
			"container", c.ID[:12], "ports", config.Ports, "elapsed", time.Since(start),
		)
		return nil
	}, m.DefaultDelay)

	return nil
}

func (m *Watcher) validateLabels(l map[string]string) error {
	if l[LabelNetworkType] == "" {
		return fmt.Errorf("invalid label %q, should be provided`", LabelNetworkType)
	}

	if l[LabelNetworkType] != "static" && l[LabelNetworkType] != "ephemeral" {
		return fmt.Errorf("invalid label %q value must be `static` or `ephemeral`", LabelNetworkType)
	}

	if l[LabelNetworkType] == "static" && l[LabelNetworkAddress] == LabelNetworkAddress {
		return fmt.Errorf("invalid label %q, cannot be empty when %q is static", LabelNetworkAddress, LabelNetworkType)
	}

	return nil
}

func (m *Watcher) createNetworkConfig(c *docker.Container, l map[string]string) *providers.NetworkConfig {
	n := m.createNetworkConfigFromLabels(l)
	n.Container = c.ID[:12]

	if c.HostConfig == nil {
		return n
	}

	for internal, externals := range c.HostConfig.PortBindings {
		for _, external := range externals {
			if external.HostIP != "0.0.0.0" && external.HostIP != "" {
				continue
			}

			n.Ports = append(n.Ports, docker.Port(external.HostPort+"/"+internal.Proto()))
		}
	}

	return n
}

func (m *Watcher) createNetworkConfigFromLabels(l map[string]string) *providers.NetworkConfig {
	n := &providers.NetworkConfig{}

	for key, value := range l {
		switch key {
		case LabelNetworkGroup:
			n.GroupName = value
		case LabelNetworkAddress:
			n.Address = value
		case LabelNetworkSourceRanges:
			n.Source.Ranges = strings.Split(value, ",")
		case LabelNetworkSourceTags:
			n.Source.Ranges = strings.Split(value, ",")
		case LabelNetworkSessionAffinity:
			n.SessionAffinity = providers.SessionAffinity(value)
		}
	}

	return n
}
