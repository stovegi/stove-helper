package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/jhump/protoreflect/dynamic"
)

func (s *Service) startHelper() error {
	http.Handle("/", http.FileServer(http.Dir("data/html")))
	http.HandleFunc("/api/tile", handleTile)
	http.HandleFunc("/api/icon", handleIcon)
	s.entityMap = make(map[uint32]*Entity)
	http.HandleFunc("/api/data/entity", func(w http.ResponseWriter, r *http.Request) {
		items := s.SelectEntity()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("["))
		for i, item := range items {
			w.Write([]byte(fmt.Sprintf("{\"id\":%d,\"data\":%s}", item.ID, item.Data)))
			if i < len(items)-1 {
				w.Write([]byte(","))
			}
		}
		w.Write([]byte("]"))
	})
	s.playerMap = make(map[uint32]*Player)
	http.HandleFunc("/api/data/player", func(w http.ResponseWriter, r *http.Request) {
		items := s.SelectPlayer()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("["))
		for i, item := range items {
			w.Write([]byte(fmt.Sprintf("{\"id\":%d,\"data\":%s}", item.ID, item.Data)))
			if i < len(items)-1 {
				w.Write([]byte(","))
			}
		}
		w.Write([]byte("]"))
	})
	return http.ListenAndServe(":8080", nil)
}

type Achievement struct {
	ID   uint32
	Data []byte
}

type Avatar struct {
	ID   uint32
	Data []byte
}

type Entity struct {
	ID   uint32
	Data []byte
}

type Player struct {
	ID   uint32
	Data []byte
}

type Item struct {
	ID   uint32
	Data []byte
}

func (s *Service) handleMessage(name string, message *Message) {
	switch name {
	case "SceneEntityAppearNotify":
		items := []*Entity{}
		for _, v := range message.GetFieldByName("entity_list").([]any) {
			v := v.(*dynamic.Message)
			item := &Entity{ID: v.GetFieldByName("entity_id").(uint32)}
			data, _ := json.Marshal(v)
			item.Data = data
			items = append(items, item)
		}
		s.UpdateEntity(items...)
	case "SceneEntityDisappearNotify":
		ids := []uint32{}
		for _, v := range message.GetFieldByName("entity_list").([]any) {
			ids = append(ids, v.(uint32))
		}
		s.DeleteEntity(ids...)
		s.DeletePlayer(ids...)
	case "CombatInvocationsNotify":
		for _, v := range message.GetFieldByName("invoke_list").([]any) {
			v := v.(*dynamic.Message)
			switch CombatTypeArgument(v.GetFieldByName("argument_type").(int32)) {
			case CombatTypeArgument_ENTITY_MOVE:
				info := NewMessage(s.protoMap["EntityMoveInfo"])
				_ = info.Unmarshal(v.GetFieldByName("combat_data").([]byte))
				data, _ := json.Marshal(info.GetFieldByName("motion_info").(*dynamic.Message))
				s.UpdatePlayer(&Player{ID: info.GetFieldByName("entity_id").(uint32), Data: data})
			}
		}
	case "ScenePlayerLocationNotify":
		items := []*Player{}
		for _, v := range message.GetFieldByName("player_loc_list").([]any) {
			v := v.(*dynamic.Message)
			item := &Player{ID: v.GetFieldByName("uid").(uint32)}
			data, _ := json.Marshal(v)
			item.Data = data
			items = append(items, item)
		}
		s.UpdatePlayer(items...)
	case "UnionCmdNotify":
		for _, v := range message.GetFieldByName("cmd_list").([]any) {
			v := v.(*dynamic.Message)
			body := NewMessage(s.protoMap[s.cmdIdMap[uint16(v.GetFieldByName("message_id").(uint32))]])
			_ = body.Unmarshal(v.GetFieldByName("body").([]byte))
			s.handleMessage(body.GetMessageDescriptor().GetName(), body)
		}
	}
}

func (s *Service) UpdateEntity(items ...*Entity) {
	s.mu.Lock()
	for _, item := range items {
		s.entityMap[item.ID] = item
	}
	s.mu.Unlock()
}

func (s *Service) DeleteEntity(ids ...uint32) {
	s.mu.Lock()
	for _, id := range ids {
		delete(s.entityMap, id)
	}
	s.mu.Unlock()
}

func (s *Service) SelectEntity() []*Entity {
	s.mu.RLock()
	items := []*Entity{}
	for _, item := range s.entityMap {
		items = append(items, item)
	}
	s.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

func (s *Service) UpdatePlayer(items ...*Player) {
	s.mu.Lock()
	for _, item := range items {
		s.playerMap[item.ID] = item
	}
	s.mu.Unlock()
}

func (s *Service) DeletePlayer(ids ...uint32) {
	s.mu.Lock()
	for _, id := range ids {
		delete(s.playerMap, id)
	}
	s.mu.Unlock()
}

func (s *Service) SelectPlayer() []*Player {
	s.mu.RLock()
	items := []*Player{}
	for _, item := range s.playerMap {
		items = append(items, item)
	}
	s.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}
