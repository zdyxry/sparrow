package pkg

import (
	"fmt"
	"strings"
	"os/exec"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
)

// NetworkConfig for link
type NetworkConfig struct {
	address *netlink.Addr
	link    netlink.Link
}

// NewNetworkConfig return NetworkConfig
func NewNetworkConfig(address, link string) (*NetworkConfig, error) {
	var networkConfig NetworkConfig
	var err error

	networkConfig.address, err = netlink.ParseAddr(address + "/32")
	if err != nil {
		err = errors.Wrapf(err, "could not parse address '%s'", address)
		return nil, err
	}

	networkConfig.link, err = netlink.LinkByName(link)
	if err != nil {
		err = errors.Wrapf(err, "could not get link for nic '%s'", link)
		return nil, err
	}

	return &networkConfig, nil
}

func (nc *NetworkConfig) IsSet() (bool, error) {
	var addresses []netlink.Addr

	addresses, err := netlink.AddrList(nc.link, 0)
	if err != nil {
		err = errors.Wrapf(err, "could not list addresses")
		return false, err
	}

	for _, addr := range addresses {
		if addr.Equal(*nc.address) {
			return true, nil
		}
	}

	return false, nil
}

func (nc *NetworkConfig) AddIP() error {
	res, err := nc.IsSet()
	if err != nil {
		return errors.Wrap(err, "ip check in AddIP failed")

	}

	if res {
		return nil
	}

	if err := netlink.AddrAdd(nc.link, nc.address); err != nil {
		return errors.Wrap(err, "could not add ip")
	}

	return nil
}

func (nc *NetworkConfig) DelIP() error {
	res , err := nc.IsSet()
	if err != nil {
		return errors.Wrap(err, "ip check in DelIP failed")
	}

	if !res {
		return nil
	}

	if err := netlink.AddrDel(nc.link, nc.address); err != nil {
		return errors.Wrap(err, "could not delete ip")
	}

	return nil
}

func (nc *NetworkConfig) IP() string {
	return nc.address.IP.String()
}

func (nc *NetworkConfig) Link() string {
	return nc.link.Attrs().Name
}

func (nc *NetworkConfig) IsUsed() bool {
	cmd := fmt.Sprintf("ping -w 1 -c 1 %s > /dev/null && echo true || echo false", nc.IP())
	output, err := exec.Command("/bin/bash", "-c", cmd).Output()
	out := strings.TrimSpace(string(output))
	fmt.Println("ping output", out)
	if err != nil || out == "false" {
		return false
	}
	return true
}