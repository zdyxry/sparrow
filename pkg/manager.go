package pkg

import (
	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type Manager struct {
	id             string
	bind           string
	fsm            FSM
	peers          RaftPeers
	logger         Logger
	stop           chan bool
	finished       chan bool
	networkManager NetworkManager
	ntpManager     NTPManager
}

func NewManager(id, bind string, peers RaftPeers, logger Logger, networkManager NetworkManager, ntpManager NTPManager) *Manager {
	return &Manager{
		id:             id,
		peers:          peers,
		bind:           bind,
		fsm:            FSM{},
		logger:         logger,
		networkManager: networkManager,
		ntpManager:     ntpManager,
	}
}

func (manager *Manager) Start() error {

	manager.stop = make(chan bool, 1)
	manager.finished = make(chan bool, 1)

	// Create config
	log.Info("Create config")
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(manager.id)
	config.LogOutput = manager.logger

	// Initialize communication
	log.Info("Initialize communication")
	address, err := net.ResolveTCPAddr("tcp", manager.bind)
	if err != nil {
		return err
	}

	// Create transport
	log.Info("Create transport")
	transport, err := raft.NewTCPTransport(manager.bind, address, 3, 10*time.Second, manager.logger)
	if err != nil {
		return err
	}

	// Create Raft structures
	log.Info("Create raft structures")
	snapshots := raft.NewInmemSnapshotStore()
	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	// Cluster configuration
	log.Info("Create raft cluster configuration")
	configuration := raft.Configuration{}
	for id, ip := range manager.peers {
		configuration.Servers = append(configuration.Servers, raft.Server{
			ID:      raft.ServerID(id),
			Address: raft.ServerAddress(ip),
		})
	}

	// Bootstrap cluster
	if err := raft.BootstrapCluster(config, logStore, stableStore, snapshots, transport, configuration); err != nil {
		return err
	}

	// Create network raft instance
	networkRaftServer, err := raft.NewRaft(config, manager.fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return err
	}
	manager.networkStart(networkRaftServer)

	// Create NTP raft instance
	ntpRaftServer, err := raft.NewRaft(config, manager.fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return err
	}
	manager.ntpStart(ntpRaftServer)

	log.Info("Started")
	return nil
}

func (manager *Manager) LeaderAddr(raftServer *raft.Raft) string {
	return string(raftServer.Leader())
}

func (manager *Manager) networkStart(raftServer *raft.Raft) {
	ticker := time.NewTicker(time.Second)
	isLeader := false

	// Delete ip when first boot
	networkManager := manager.networkManager
	var leaderIP string
	networkManager.DelIP()
	go func() {
		for {
			select {
			case leader := <-raftServer.LeaderCh():
				if leader {
					isLeader = true
					log.Info("Network leading")
					networkManager.AddIP()
				}
			case <-ticker.C:
				if isLeader {
					result, err := manager.networkManager.IsSet()
					if err != nil {
						log.WithFields(log.Fields{
							"error": err,
							"ip":    manager.networkManager.IP(),
							"link":  manager.networkManager.Link(),
						}).Error("Could not check ip")
					}

					if result == false {
						log.Error("Lost IP")
						networkManager.AddIP()
					}
				} else {
					curLeaderIP := manager.LeaderAddr(raftServer)
					if curLeaderIP != "" && leaderIP != curLeaderIP {
						log.WithField("Leader Address", curLeaderIP).Info("Network Current Leader IP")
						leaderIP = curLeaderIP
					}
				}
			case <-manager.stop:
				log.Info("Stopping")
				if isLeader {
					networkManager.DelIP()
				}
				close(manager.finished)
				return
			}
		}
	}()
}

func (manager *Manager) ntpStart(raftServer *raft.Raft) {
	var leaderIP string
	ticker := time.NewTicker(time.Second)
	isLeader := false
	ntpManager := manager.ntpManager

	go func() {
		for {
			select {
			case leader := <-raftServer.LeaderCh():
				if leader {
					isLeader = true
					log.Info("NTP leading")
					ntpManager.RenderLeader()
					ntpManager.RestartService()
				}
			case <-ticker.C:
				if !isLeader {
					curLeaderIP := manager.LeaderAddr(raftServer)
					if curLeaderIP != "" && leaderIP != curLeaderIP {
						log.WithField("Leader Address", curLeaderIP).Info("NTP current Leader IP")
						leaderIP = curLeaderIP
						ip, err := SplitIP(curLeaderIP)
						if err != nil {
							log.Error("Leader Address is invalid")
						}
						ntpManager.RenderFollower(ip)
						ntpManager.RestartService()
					}

				}
			}
		}
	}()
}

func (manager *Manager) Stop() {
	close(manager.stop)
	<-manager.finished
	log.Info("Manager stopped")
}
