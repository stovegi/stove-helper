package helper

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/rs/zerolog/log"
)

type Message struct{ *dynamic.Message }

func NewMessage(md *desc.MessageDescriptor) *Message {
	return &Message{dynamic.NewMessage(md)}
}

func (m *Message) MarshalJSON() ([]byte, error) {
	return m.MarshalJSONPB(&jsonpb.Marshaler{OrigName: true})
}

type Packet struct {
	flag bool
	data []byte
	time time.Time
	head *Message
	body *Message
}

func (s *Service) handlePacket(packet *Packet) {
	s.xor(packet.data)
	p := packet.data
	l := len(p)
	if p[0] != 0x45 || p[1] != 0x67 || p[l-2] != 0x89 || p[l-1] != 0xAB {
		log.Warn().Uint16("magic", binary.BigEndian.Uint16(p)).Int64("ts", packet.time.UnixMilli()).Msg("Invalid packet, maybe not encrypted")
		return
	}
	headLength := binary.BigEndian.Uint16(p[4:])
	bodyLength := binary.BigEndian.Uint32(p[6:])
	packet.head = NewMessage(s.protos["PacketHead"])
	_ = packet.head.Unmarshal(p[10 : 10+headLength])
	packet.body = NewMessage(s.protos[s.cmdIds[binary.BigEndian.Uint16(p[2:])]])
	_ = packet.body.Unmarshal(p[10+headLength : 10+uint32(headLength)+bodyLength])
	s.onPacket(packet)
}

type UnionCmdNotify struct {
	CmdList []*UnionCmd `json:"cmd_list"`
}

type UnionCmd struct {
	MessageID uint16   `json:"message_id"`
	Body      *Message `json:"body"`
}

func (s *Service) onPacket(packet *Packet) {
	seq, name := packet.head.GetFieldByName("client_sequence_id").(uint32), packet.body.GetMessageDescriptor().GetName()
	info := ""
	if packet.flag {
		log.Info().Uint32("#seq", seq).Str("cmd", name).Msg("SERVER > Incoming > CLIENT")
		info = fmt.Sprintf(`"SERVER > Incoming > CLIENT #seq=%d #time=%d cmd=%s"`, seq, packet.time.UnixMilli(), name)
	} else {
		log.Info().Uint32("#seq", seq).Str("cmd", name).Msg("CLIENT > Outgoing > SERVER")
		info = fmt.Sprintf(`"CLIENT > Outgoing > SERVER #seq=%d #time=%d cmd=%s"`, seq, packet.time.UnixMilli(), name)
	}
	headJson, _ := json.Marshal(packet.head)
	bodyJson, _ := json.Marshal(packet.body)
	if name == "GetPlayerTokenReq" {
		s.sentMs = packet.head.GetFieldByName("sent_ms").(uint64)
		log.Info().Uint64("sent_ms", s.sentMs).Msg("Sent ms")
	} else if name == "GetPlayerTokenRsp" {
		serverRandKey := packet.body.GetFieldByName("server_rand_key").(string)
		seed, err := base64.StdEncoding.DecodeString(serverRandKey)
		if err != nil {
			log.Error().Err(err).Msg("Failed to decode server rand key")
			return
		}
		seed, err = decrypt(s.priv, seed)
		if err != nil {
			log.Error().Err(err).Msg("Failed to decrypt server rand key")
			return
		}
		s.serverSeed = binary.BigEndian.Uint64(seed)
		log.Info().Uint64("seed", s.serverSeed).Msg("Server seed")
	} else if name == "UnionCmdNotify" {
		notify := &UnionCmdNotify{}
		for _, v := range packet.body.GetFieldByName("cmd_list").([]any) {
			v := v.(*dynamic.Message)
			item := &UnionCmd{
				MessageID: uint16(v.GetFieldByName("message_id").(uint32)),
			}
			item.Body = NewMessage(s.protos[s.cmdIds[item.MessageID]])
			_ = item.Body.Unmarshal(v.GetFieldByName("body").([]byte))
			notify.CmdList = append(notify.CmdList, item)
		}
		bodyJson, _ = json.Marshal(notify)
	}
	fmt.Fprintf(s.file, "- info: %s\n  head: %s\n  body: %s\n", info, headJson, bodyJson)
	s.file.Sync()
	s.handleMessage(name, packet.body)
}
