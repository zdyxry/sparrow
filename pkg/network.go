package pkg

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type NetworkManager interface {
	AddIP()
	DelIP()
	IsSet() (bool, error)
	IsUsed() bool
	IP() string
	Link() string
	SendARP() error
}

// NetworkConfig for link
type NetworkConfig struct {
	address *netlink.Addr
	link    netlink.Link
}

// NewNetworkConfig return NetworkConfig
func NewNetworkConfig(address, link string) (*NetworkConfig, error) {
	var networkConfig NetworkConfig
	var err error

	networkConfig.address, err = netlink.ParseAddr(address + "/24")
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

func (nc *NetworkConfig) AddIP() {
	if err := nc.addIP(); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"ip":    nc.IP(),
			"link":  nc.Link(),
		}).Error("Could not set ip")
	} else {
		log.WithFields(logrus.Fields{
			"ip":   nc.IP(),
			"link": nc.Link(),
		}).Info("Add IP success")
	}
}

func (nc *NetworkConfig) DelIP() {
	if err := nc.delIP(); err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"ip":    nc.IP(),
			"link":  nc.Link(),
		}).Error("Could not delete ip")
	} else {
		log.WithFields(logrus.Fields{
			"ip":   nc.IP(),
			"link": nc.Link(),
		}).Info("Delete IP success")
	}
}

func (nc *NetworkConfig) delIP() error {
	res, err := nc.IsSet()
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
	if err != nil || out == "false" {
		return false
	}
	return true
}

func (nc *NetworkConfig) addIP() error {
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

func (nc *NetworkConfig) SendARP() error {
	var arp *ARPGratuitous
	arp = &ARPGratuitous{
		IfaceName: nc.Link(),
		IP: net.ParseIP(nc.IP()),
	}
	if err := ARPSendGratuitous(arp); err != nil {
		return err
	}
	return nil
}
