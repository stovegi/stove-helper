package helper

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/google/gopacket/pcap"
	"github.com/jhump/protoreflect/desc"
	"github.com/stovegi/stove-helper/pkg/config"
	"github.com/stovegi/stove-helper/pkg/kcp"
)

type Service struct {
	ctx context.Context
	cfg *config.Config

	// sniffer related fields
	file   *os.File
	handle *pcap.Handle

	priv     *rsa.PrivateKey
	keyStore *KeyStore
	cmdIds   map[uint16]string
	protos   map[string]*desc.MessageDescriptor

	sentMs     uint64
	serverSeed uint64
	incoming   *kcp.ControlBlock
	outgoing   *kcp.ControlBlock

	// helper related fields
	dataMap map[uint32]*Data

	mu             sync.RWMutex
	achievementMap map[uint32]*Achievement
	avatarMap      map[uint32]*Avatar
	entityMap      map[uint32]*Entity
	playerMap      map[uint32]*Player
	itemMap        map[uint32]*Item
}

func NewService(ctx context.Context, cfg config.Config) (*Service, error) {
	s := &Service{ctx: ctx, cfg: &cfg}
	if err := s.initSniffer(); err != nil {
		return nil, err
	}
	if err := s.initHelper(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Service) Start() error {
	go s.startSniffer()
	return s.startHelper()
}

func (s *Service) cacheGet(url string) ([]byte, error) {
	name := path.Join("data/cache", strings.NewReplacer(":", "-", "/", "-", "?", "-").Replace(url))
	body, err := os.ReadFile(name)
	if err == nil {
		return body, nil
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status is not ok")
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(name, body, 0644); err != nil {
		return nil, err
	}
	return body, nil
}
