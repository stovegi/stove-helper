package helper

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/rs/zerolog/log"

	"github.com/stovegi/stove-helper/pkg/ec2b"
	"github.com/stovegi/stove-helper/pkg/kcp"
)

func decrypt(priv *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	out := make([]byte, 0, 1024)
	for len(ciphertext) > 0 {
		chunkSize := 256
		if chunkSize > len(ciphertext) {
			chunkSize = len(ciphertext)
		}
		chunk := ciphertext[:chunkSize]
		ciphertext = ciphertext[chunkSize:]
		b, err := rsa.DecryptPKCS1v15(rand.Reader, priv, chunk)
		if err != nil {
			return nil, err
		}
		out = append(out, b...)
	}
	return out, nil
}

func (s *Service) parseLines(parser protoparse.Parser, lines ...string) error {
	for _, line := range lines {
		parts := strings.Split(strings.TrimSpace(line), ",")
		if len(parts) > 2 {
			continue
		}
		if parts[0] == "DebugNotify" {
			continue
		}
		if len(parts) == 2 {
			v, err := strconv.ParseUint(parts[1], 10, 16)
			if err != nil {
				return err
			}
			s.cmdIds[uint16(v)] = parts[0]
		}
		dsec, err := parser.ParseFiles(parts[0] + ".proto")
		if err != nil {
			return err
		}
		s.protos[parts[0]] = dsec[0].FindMessage(parts[0])
	}
	return nil
}

func (s *Service) parseSecret(url string) error {
	log.Info().Str("url", url).Msg("Initializing secret")
	body, err := s.cacheGet(url)
	if err != nil {
		return err
	}
	var v map[string]string
	if err := json.Unmarshal(body, &v); err != nil {
		return err
	}
	content, err := base64.StdEncoding.DecodeString(v["content"])
	if err != nil {
		return err
	}
	body, err = decrypt(s.priv, content)
	if err != nil {
		return err
	}
	pb := dynamic.NewMessage(s.protos["QueryCurrRegionHttpRsp"])
	if err := pb.Unmarshal(body); err != nil {
		return err
	}
	ec2b, err := ec2b.Load(pb.GetFieldByName("client_secret_key").([]byte))
	if err != nil {
		return err
	}
	key := ec2b.Key()
	s.keyStore.keyMap[binary.BigEndian.Uint16(key)^0x4567] = key
	log.Info().Msg("Successfully initialized secret")
	return nil
}

func (s *Service) initSniffer() error {
	// Parsing protos
	s.cmdIds = make(map[uint16]string)
	s.protos = make(map[string]*desc.MessageDescriptor)

	p, err := os.ReadFile(s.cfg.DataConfig.CmdIDPath)
	if err != nil {
		return err
	}
	prsr := protoparse.Parser{ImportPaths: []string{s.cfg.DataConfig.ProtoPath}}
	if err := s.parseLines(prsr, "QueryCurrRegionHttpRsp", "PacketHead", "EntityMoveInfo"); err != nil {
		return err
	}
	if err := s.parseLines(prsr, strings.Split(string(p), "\n")...); err != nil {
		return err
	}
	log.Info().Int("#ids", len(s.cmdIds)).Int("#protos", len(s.protos)).Msg("Successfully parsed protos")

	// Parsing secret
	s.keyStore = &KeyStore{keyMap: make(map[uint16][]byte)}

	rest, _ := os.ReadFile(s.cfg.DataConfig.PrivateKeyPath)
	var ok bool
	var block *pem.Block
	for {
		block, rest = pem.Decode(rest)
		if block.Type == "RSA PRIVATE KEY" {
			k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return err
			} else if s.priv, ok = k.(*rsa.PrivateKey); !ok {
				return fmt.Errorf("failed to parse private key")
			}
			break
		}
		if len(rest) == 0 {
			if s.priv == nil {
				return fmt.Errorf("failed to parse private key")
			}
			break
		}
	}
	if err = s.parseSecret(s.cfg.DataConfig.DispatchRegion); err != nil {
		return err
	}
	log.Info().Msg("Successfully parsed secret")

	// Open pcap handle
	s.handle, err = pcap.OpenLive(s.cfg.Device, 1600, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	if err := s.handle.SetBPFFilter("udp portrange 22101-22102"); err != nil {
		return err
	}
	log.Info().Str("device", s.cfg.Device).Msg("Successfully opened pcap handle")
	return nil
}

func (s *Service) startSniffer() {
	pcap, err := os.Create(path.Join(s.cfg.DataConfig.OutputPath, time.Now().Format("2006-01-02 15-04-05")+".pcap"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create pcap file")
		return
	}
	defer pcap.Close()
	s.file, err = os.Create(path.Join(s.cfg.DataConfig.OutputPath, time.Now().Format("2006-01-02 15-04-05")+".yaml"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create yaml file")
		return
	}
	defer s.file.Close()
	w, err := pcapgo.NewNgWriter(pcap, s.handle.LinkType())
	if err != nil {
		log.Error().Err(err).Msg("Failed to create pcap writer")
		return
	}
	ps := gopacket.NewPacketSource(s.handle, s.handle.LinkType())
	for packet := range ps.Packets() {
		packet.Metadata().CaptureInfo.InterfaceIndex = 0 // fix interface index error
		err := w.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
		if err != nil {
			log.Error().Err(err).Msg("Failed to write packet to pcap")
			return
		}
		p := packet.ApplicationLayer().Payload()
		if len(p) < kcp.IKCP_OVERHEAD {
			continue
		}
		udp := packet.TransportLayer().(*layers.UDP)
		s.handlePayload(p, udp.SrcPort == 22101 || udp.SrcPort == 22102, packet.Metadata().Timestamp)
	}
}

func (s *Service) handlePayload(p []byte, flag bool, t time.Time) {
	conv := binary.LittleEndian.Uint32(p)
	token := binary.LittleEndian.Uint32(p[4:])
	var cb *kcp.ControlBlock
	if flag {
		if s.incoming == nil || s.incoming.Conv() != conv || s.incoming.Token() != token {
			log.Info().Uint32("conv", conv).Uint32("token", token).Msg("New incoming kcp session")
			s.incoming = kcp.NewControlBlock(conv, kcp.NilOutputFunc)
			s.incoming.SetToken(token)
			s.incoming.SetMtu(1200)
			s.incoming.NoDelay(1, 20, 2, 1)
			s.incoming.WndSize(255, 255)
		}
		cb = s.incoming
	} else {
		if s.outgoing == nil || s.outgoing.Conv() != conv || s.outgoing.Token() != token {
			log.Info().Uint32("conv", conv).Uint32("token", token).Msg("New outgoing kcp session")
			s.outgoing = kcp.NewControlBlock(conv, kcp.NilOutputFunc)
			s.outgoing.SetToken(token)
			s.outgoing.SetMtu(1200)
			s.outgoing.NoDelay(1, 20, 2, 1)
			s.outgoing.WndSize(255, 255)
		}
		cb = s.outgoing
	}
	_ = cb.Input(p, true, true)
	size := cb.PeekSize()
	for size > 0 {
		packet := &Packet{}
		packet.flag = flag
		packet.data = make([]byte, size)
		packet.time = t
		_ = cb.Recv(packet.data)
		go s.handlePacket(packet)
		size = cb.PeekSize()
	}
	cb.Update()
}
