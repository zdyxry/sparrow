package pkg

import (
	"github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

var (
	leaderIP    string
	curLeaderIP string
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
	raftServer, err := raft.NewRaft(config, manager.fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return err
	}
	manager.Action(raftServer)

	log.Info("Started")
	return nil
}

func (manager *Manager) LeaderAddr(raftServer *raft.Raft) string {
	return string(raftServer.Leader())
}

func (manager *Manager) Action(raftServer *raft.Raft) {
	ticker := time.NewTicker(time.Second)
	isLeader := false

	// Delete ip when first boot
	networkManager := manager.networkManager
	networkManager.DelIP()
	go func() {
		for {
			select {
			case leader := <-raftServer.LeaderCh():
				if leader {
					isLeader = true
					log.Info("Network leading")
					manager.leaderAction()
				}
			case <-ticker.C:
				if raftServer.State() == raft.RaftState(uint32(2)) {
					manager.leaderCheck()
				} else {
					manager.followerAction(raftServer)
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

func (manager *Manager) leaderAction() {
	networkManager := manager.networkManager
	ntpManager := manager.ntpManager
	go func() {
		networkManager.AddIP()
		if err := networkManager.SendARP(); err != nil {
			log.Error(err)
		}
	}()

	go func() {
		ntpManager.RenderLeader()
		ntpManager.RestartService()
	}()
}

func (manager *Manager) followerAction(raftServer *raft.Raft) {
	curLeaderIP := manager.LeaderAddr(raftServer)
	if curLeaderIP != "" && leaderIP != curLeaderIP {
		log.WithField("Leader Address", curLeaderIP).Info("NTP current Leader IP")
		leaderIP = curLeaderIP
		ip, err := SplitIP(curLeaderIP)
		if err != nil {
			log.Error("Leader Address is invalid")
		}
		manager.ntpManager.RenderFollower(ip)
		manager.ntpManager.RestartService()
	}
}

func (manager *Manager) leaderCheck() {
	result, err := manager.networkManager.IsSet()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"ip":    manager.networkManager.IP(),
			"link":  manager.networkManager.Link(),
		}).Error("Could not check ip")
	}

	if result == false {
		log.Error("Lost IP")
		manager.networkManager.AddIP()
	}
}

func (manager *Manager) Stop() {
	close(manager.stop)
	<-manager.finished
	log.Info("Manager stopped")
}
