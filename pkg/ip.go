package pkg

import (
	"fmt"
	"net"
	"strings"
)

func SplitIP(addr string) (string, error) {
	var err error
	res := strings.Split(addr, ":")
	ip := res[0]
	if ip := net.ParseIP(ip); ip != nil {
		return ip.String(), nil
	} else {
		err = fmt.Errorf("Failed to parse IP address")
		return "", err
	}
}
