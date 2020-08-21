package docker

import (
	"fmt"
	"time"

	"github.com/CodyGuo/glog"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
)

func (c *Container) EventStart(h func(Container) error) <-chan error {
	errStream := make(chan error)
	c.event.Handle("start", func(e events.Message) {
		name, err := c.getName(e.ID)
		if err != nil {
			errStream <- err
		}
		sandboxKey, err := c.getSandboxKey(e.ID)
		if err != nil {
			errStream <- err
			return
		}
		veths, err := c.GetVeths(name, sandboxKey)
		if err != nil {
			errStream <- err
			return
		}
		rate, ceil := c.getLabelTC(e.Actor.Attributes)
		for _, veth := range veths {
			err = h(Container{
				ID:     e.ID[:12],
				Name:   name,
				Veth:   veth,
				TcRate: rate,
				TcCeil: ceil,
			})
		}
		if err != nil {
			errStream <- err
		}
	})
	return errStream
}

func (c *Container) EventDie(h func(Container) error) <-chan error {
	errStream := make(chan error)
	c.event.Handle("die", func(e events.Message) {
		name, err := c.getName(e.ID)
		if err != nil {
			errStream <- fmt.Errorf("getName error: %w", err)
		}
		err = h(Container{
			ID:   e.ID[:12],
			Name: name,
		})
		if err != nil {
			errStream <- err
		}
	})
	return errStream
}

func (c *Container) eventWatch() {
	eventStream := make(chan events.Message)
	go func() {
		f := filters.NewArgs()
		f.Add("type", "container")
		f.Add("label", "org.label-schema.tc.enabled=1")
		eventMsg, eventErr := c.dc.Events(c.ctx, types.EventsOptions{Filters: f})
		for {
			select {
			case err := <-eventErr:
				glog.Errorf("eventWatch failed, error: %v, Try again after 5 seconds", err)
				time.Sleep(5 * time.Second)
				eventMsg, eventErr = c.dc.Events(c.ctx, types.EventsOptions{Filters: f})
			case msg := <-eventMsg:
				eventStream <- msg
			}
		}
	}()
	c.event = InitEventHandler()
	c.event.Watch(eventStream)
}
