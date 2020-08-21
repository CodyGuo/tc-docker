package docker

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/CodyGuo/glog"
	"github.com/CodyGuo/tc-docker/pkg/command"
)

type Veth struct {
	Device    string
	Ident     string
	LinkIdent string
}

func (c *Container) GetVeths(name, sandboxKey string) ([]string, error) {
	containerVeths, err := c.getContainerVeths(name, sandboxKey)
	if err != nil {
		return nil, err
	}
	hostVeths, err := c.getHostVeths()
	if err != nil {
		return nil, err
	}
	veths := []string{}
	for _, hv := range hostVeths {
		for _, cv := range containerVeths {
			if cv.Ident == hv.LinkIdent && cv.LinkIdent == hv.Ident {
				glog.Debugf("GetVeths found, container: %s, device: %s, veth: %+v", name, hv.Device, *cv)
				veths = append(veths, hv.Device)
			}
		}
	}
	if len(veths) == 0 {
		return nil, fmt.Errorf("container: %s, not found veth", name)
	}
	return veths, nil
}

func (c *Container) RemoveVeth(name string) error {
	veth := "/var/run/netns/" + name
	glog.Debugf("RemoveVeth: %s", veth)
	return os.Remove(veth)
}

func (c *Container) getHostVeths() ([]*Veth, error) {
	ipAddrCmd := "/usr/sbin/ip addr show type veth"
	glog.Debug(ipAddrCmd)
	out, err := command.CombinedOutput(ipAddrCmd)
	if err != nil {
		return nil, fmt.Errorf("out: %s, error: %v", out, err)
	}
	var veths []*Veth
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		b := scanner.Bytes()
		if !bytes.Contains(b, []byte("veth")) {
			continue
		}
		veth, err := parseVeth(b)
		if err != nil {
			glog.Errorf("getAllVeth, parseVeth: %v, veth: %s", err, b)
			continue
		}
		veths = append(veths, &veth)
	}
	return veths, nil
}

func (c *Container) getContainerVeths(name, sandboxKey string) ([]*Veth, error) {
	os.Remove("/var/run/netns/" + name)
	if err := os.Symlink(sandboxKey, "/var/run/netns/"+name); err != nil {
		return nil, err
	}
	ipAddrCmd := fmt.Sprintf("/usr/sbin/ip netns exec %s ip addr show ", name)
	glog.Debug(ipAddrCmd)
	out, err := command.CombinedOutput(ipAddrCmd)
	if err != nil {
		return nil, fmt.Errorf("out: %s, error: %v", out, err)
	}
	var veths []*Veth
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		b := scanner.Bytes()
		if !bytes.Contains(b, []byte("UP")) {
			continue
		}
		if bytes.Contains(b, []byte("LOOPBACK")) {
			continue
		}
		veth, err := parseVeth(b)
		if err != nil {
			glog.Errorf("getContainerVeth, parseVeth: %v, veth: %s", err, b)
			continue
		}
		veths = append(veths, &veth)
	}
	return veths, nil
}

func parseVeth(b []byte) (Veth, error) {
	fields := bytes.Split(b, []byte(":"))
	if len(fields) < 2 {
		return Veth{}, errors.New("not found")
	}
	ident := bytes.TrimSpace(fields[0])
	devices := bytes.Split(bytes.TrimSpace(fields[1]), []byte("@if"))
	if len(devices) < 2 {
		return Veth{}, errors.New("not found")

	}
	device := devices[0]
	link := devices[1]
	glog.Debugf("parseVeth, ident: %s, device: %s, link: %s", ident, device, link)
	return Veth{
		Device:    string(devices[0]),
		Ident:     string(ident),
		LinkIdent: string(devices[1]),
	}, nil
}
