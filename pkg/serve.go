package pkg

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

func Serve(cfg Config) {
	// NetworkManager
	netcfg := cfg.Network
	networkManager, err := NewNetworkConfig(netcfg.Vip, netcfg.Interface)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Init networkManager failed")
		os.Exit(1)
	}

	// NTPManager
	ntpcfg := cfg.Ntp
	var isInternal bool
	if ntpcfg.Mode == "internal" {
		isInternal = true
	} else if ntpcfg.Mode == "external" {
		isInternal = false
	}
	ntpManager, err := NewNTPConfig(isInternal, ntpcfg.Servers)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Init ntpManager failed")
		os.Exit(1)
	}

	// Config Raft peers
	raftcfg := cfg.Raft
	peers := RaftPeers{}
	if len(raftcfg.Peers) > 0 {
		for _, peer := range raftcfg.Peers {
			peers[peer.ID] = peer.Addr
		}
	}

	// Manager Start
	logger := Logger{}
	manager := NewManager(raftcfg.ID, raftcfg.Bind, peers, logger, networkManager, ntpManager)
	if err := manager.Start(); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Manager start failed")
		os.Exit(1)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	manager.Stop()
}
