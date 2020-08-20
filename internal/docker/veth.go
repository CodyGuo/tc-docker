package docker

import (
	"bufio"
	"bytes"
	"errors"
	"os"

	"github.com/CodyGuo/glog"
	"github.com/CodyGuo/tc-docker/pkg/command"
)

type Veth struct {
	Device    string
	Ident     string
	LinkIdent string
}

func (c *Container) getVeth(name, sandboxKey string) (string, error) {
	veth, err := c.getContanierVeth(name, sandboxKey)
	if err != nil {
		return "", err
	}
	veths, err := c.getAllVeth()
	if err != nil {
		return "", err
	}
	for _, v := range veths {
		if veth.Ident == v.LinkIdent && veth.LinkIdent == v.Ident {
			glog.Debugf("getVeth, container: %s, device: %s", name, v.Device)
			return string(v.Device), nil
		}
	}
	return "", errors.New("not found veth")
}

func (c *Container) getAllVeth() ([]*Veth, error) {
	ethxs, err := command.CombinedOutput("/usr/sbin/ip addr")
	if err != nil {
		return nil, err
	}
	var veths []*Veth
	scanner := bufio.NewScanner(bytes.NewReader(ethxs))
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

func (c *Container) getContanierVeth(name, sandboxKey string) (Veth, error) {
	os.Remove("/var/run/netns/" + name)
	os.Symlink(sandboxKey, "/var/run/netns/"+name)
	ethxs, err := command.CombinedOutput("/usr/sbin/ip netns exec " + name + " ip addr show eth0")
	if err != nil {
		return Veth{}, err
	}
	return parseVeth(ethxs)
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
