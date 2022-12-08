package helper

import (
	"crypto/rsa"
	"os"
	"sync"

	"github.com/StoveGI/stove-helper/pkg/config"
	"github.com/StoveGI/stove-helper/pkg/net"
	"github.com/google/gopacket/pcap"
	"github.com/jhump/protoreflect/desc"
)

type Service struct {
	config *config.Config

	// sniffer related
	rawlog *os.File
	handle *pcap.Handle

	priv     *rsa.PrivateKey
	keyStore *KeyStore
	cmdIdMap map[uint16]string
	protoMap map[string]*desc.MessageDescriptor

	sentMs     uint64
	serverSeed uint64
	incoming   *net.KCP
	outgoing   *net.KCP

	// helper related
	mu             sync.RWMutex
	achievementMap map[uint32]*Achievement
	avatarMap      map[uint32]*Avatar
	entityMap      map[uint32]*Entity
	playerMap      map[uint32]*Player
	itemMap        map[uint32]*Item
}

func NewService(c config.Config) (*Service, error) {
	s := &Service{config: &c}
	if err := s.initSniffer(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Service) Start() error {
	go s.startSniffer()
	return s.startHelper()
}
