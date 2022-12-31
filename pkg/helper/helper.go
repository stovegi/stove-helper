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
	http.HandleFunc("/api/icon", s.handleIcon)
	s.achievementMap = make(map[uint32]*Achievement)
	http.HandleFunc("/api/data/achievement", func(w http.ResponseWriter, r *http.Request) {
		items := s.SelectAchievement()
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
	s.avatarMap = make(map[uint32]*Avatar)
	http.HandleFunc("/api/data/avatar", func(w http.ResponseWriter, r *http.Request) {
		items := s.SelectAvatar()
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
	s.itemMap = make(map[uint32]*Item)
	http.HandleFunc("/api/data/item", func(w http.ResponseWriter, r *http.Request) {
		items := s.SelectItem()
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

func (s *Service) UpdateAchievement(items ...*Achievement) {
	s.mu.Lock()
	for _, item := range items {
		s.achievementMap[item.ID] = item
	}
	s.mu.Unlock()
}

func (s *Service) SelectAchievement() []*Achievement {
	s.mu.RLock()
	items := []*Achievement{}
	for _, item := range s.achievementMap {
		items = append(items, item)
	}
	s.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

type Avatar struct {
	ID   uint32
	Data []byte
}

func (s *Service) UpdateAvatar(items ...*Avatar) {
	s.mu.Lock()
	for _, item := range items {
		s.avatarMap[item.ID] = item
	}
	s.mu.Unlock()
}

func (s *Service) SelectAvatar() []*Avatar {
	s.mu.RLock()
	items := []*Avatar{}
	for _, item := range s.avatarMap {
		items = append(items, item)
	}
	s.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

type Entity struct {
	ID   uint32
	Data []byte
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

type Player struct {
	ID   uint32
	Data []byte
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

type Item struct {
	ID   uint32
	Data []byte
}

func (s *Service) UpdateItem(items ...*Item) {
	s.mu.Lock()
	for _, item := range items {
		s.itemMap[item.ID] = item
	}
	s.mu.Unlock()
}

func (s *Service) DeleteItem(ids ...uint32) {
	s.mu.Lock()
	for _, id := range ids {
		delete(s.itemMap, id)
	}
	s.mu.Unlock()
}

func (s *Service) SelectItem() []*Item {
	s.mu.RLock()
	items := []*Item{}
	for _, item := range s.itemMap {
		items = append(items, item)
	}
	s.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

func (s *Service) handleMessage(name string, message *Message) {
	switch name {
	case "AchievementAllDataNotify", "AchievementUpdateNotify":
		items := []*Achievement{}
		for _, v := range message.GetFieldByName("achievement_list").([]any) {
			v := v.(*dynamic.Message)
			item := &Achievement{ID: v.GetFieldByName("id").(uint32)}
			data, _ := json.Marshal(v)
			item.Data = data
			items = append(items, item)
		}
		s.UpdateAchievement(items...)
	case "AvatarDataNotify":
		items := []*Avatar{}
		for _, v := range message.GetFieldByName("avatar_list").([]any) {
			v := v.(*dynamic.Message)
			item := &Avatar{ID: uint32(v.GetFieldByName("guid").(uint64))}
			data, _ := json.Marshal(v)
			item.Data = data
			items = append(items, item)
		}
		s.UpdateAvatar(items...)
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
		fmt.Println("SceneEntityDisappearNotify", message)
		ids := []uint32{}
		for _, v := range message.GetFieldByName("entity_list").([]any) {
			ids = append(ids, v.(uint32))
		}
		switch VisionType(message.GetFieldByName("disappear_type").(int32)) {
		default:
			s.DeleteEntity(ids...)
		case VisionType_VISION_REPLACE:
			s.DeleteEntity(ids...)
			s.DeletePlayer(ids...)
		case VisionType_VISION_MISS:
		}
	case "CombatInvocationsNotify":
		for _, v := range message.GetFieldByName("invoke_list").([]any) {
			v := v.(*dynamic.Message)
			switch CombatTypeArgument(v.GetFieldByName("argument_type").(int32)) {
			case CombatTypeArgument_ENTITY_MOVE:
				info := NewMessage(s.protos["EntityMoveInfo"])
				_ = info.Unmarshal(v.GetFieldByName("combat_data").([]byte))
				data, _ := json.Marshal(info.GetFieldByName("motion_info").(*dynamic.Message))
				switch id := info.GetFieldByName("entity_id").(uint32); id >> 24 {
				default:
					s.UpdateEntity(&Entity{ID: id, Data: data})
				case 1:
					s.UpdatePlayer(&Player{ID: id, Data: data})
				}
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
			body := NewMessage(s.protos[s.cmdIds[uint16(v.GetFieldByName("message_id").(uint32))]])
			_ = body.Unmarshal(v.GetFieldByName("body").([]byte))
			s.handleMessage(body.GetMessageDescriptor().GetName(), body)
		}
	case "PlayerStoreNotify", "StoreItemChangeNotify":
		items := []*Item{}
		for _, v := range message.GetFieldByName("item_list").([]any) {
			v := v.(*dynamic.Message)
			item := &Item{ID: uint32(v.GetFieldByName("guid").(uint64))}
			data, _ := json.Marshal(v)
			item.Data = data
			items = append(items, item)
		}
		s.UpdateItem(items...)
	}
}
