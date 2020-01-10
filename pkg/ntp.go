package pkg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/pkg/errors"
)

type NTPManager interface {
	RenderLeader() error
	RenderFollower(leaderIP string) error
	RestartService() error
}


type NTPConfig struct {
	LeaderIP   string
	IsInternal bool
	NTPServers []string
}

func NewNTPConfig(isInternal bool, ntpServers []string) (*NTPConfig, error) {
	var ntpConfig NTPConfig
	var err error

	for _, server := range ntpServers {
		addr := net.ParseIP(server)
		if addr == nil {
			err = errors.Wrapf(err, "failed to parse NTP Server '%s'", server)
			return nil, err
		}
	}

	ntpConfig.IsInternal = isInternal
	ntpConfig.NTPServers = ntpServers

	return &ntpConfig, nil
}

func (ntpConfig *NTPConfig) RenderLeader() error {
	ntpConfig.LeaderIP = ""
	if ntpConfig.IsInternal {
		ntpConfig.NTPServers = []string{}
	}
	log.Info("Render leader config file: %s", ntpConfig)
	return renderFile(chronyLeader, ntpConfig)
}

func (ntpConfig *NTPConfig) RenderFollower(leaderIP string) error {
	if leaderIP == "" {
		var err error
		return errors.Wrapf(err, "LeaderIP is '%s'", leaderIP)
	}
	ntpConfig.LeaderIP = leaderIP
	log.Info("Render follower config file: %s", ntpConfig)
	return renderFile(chronyFollower, ntpConfig)
}

func (ntpConfig *NTPConfig) RestartService() error {
	return restartService(chronyServiceName)
}

func run(command string, arguments ...string) error {
	cmd := exec.Command(command, arguments...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%q failed: %v, %s", command, err, out)
	}
	return nil
}

func restartService(serviceName string) error {
	err := run("systemctl", "daemon-reload")
	if err != nil {
		return errors.Wrapf(err, "failed to restart service %s", serviceName)
	}
	return run("systemctl", "restart", serviceName)
}

func renderFile(tmpl string, ntpConfig *NTPConfig) error {
	b := bytes.Buffer{}
	t := template.Must(template.New("").Parse(tmpl))
	t.Execute(&b, ntpConfig)
	if err := ioutil.WriteFile(chronyConfigFile, b.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}
