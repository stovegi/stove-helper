package helper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func (s *Service) initHelper() (err error) {
	s.dataMap = make(map[uint32]*Data)
	if err = s.initData("https://api.ambr.top/v2/en/avatar", s.dataMap); err != nil {
		return err
	}
	if err = s.initData("https://api.ambr.top/v2/en/weapon", s.dataMap); err != nil {
		return err
	}
	if err = s.initData("https://api.ambr.top/v2/en/food", s.dataMap); err != nil {
		return err
	}
	if err = s.initData("https://api.ambr.top/v2/en/material", s.dataMap); err != nil {
		return err
	}
	if err = s.initData("https://api.ambr.top/v2/en/furniture", s.dataMap); err != nil {
		return err
	}
	if err = s.initData("https://api.ambr.top/v2/en/reliquary", s.dataMap); err != nil {
		return err
	}
	if err = s.initData("https://api.ambr.top/v2/en/monster", s.dataMap); err != nil {
		return err
	}
	if err = s.initData("https://api.ambr.top/v2/en/book", s.dataMap); err != nil {
		return err
	}
	return nil
}

type Data struct {
	ID   uint32
	Name string
	Icon string
}

type Uint32orString struct {
	u uint32
	s string
}

func (u *Uint32orString) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		return json.Unmarshal(b, &u.s)
	}
	return json.Unmarshal(b, &u.u)
}

func (u Uint32orString) MarshalJSON() ([]byte, error) {
	if u.s != "" {
		return json.Marshal(u.s)
	}
	return json.Marshal(u.u)
}

func (u Uint32orString) String() string {
	if u.s != "" {
		return u.s
	}
	return fmt.Sprint(u.u)
}

func (u Uint32orString) Uint32() uint32 {
	if u.s != "" {
		u, err := strconv.ParseUint(strings.SplitN(u.s, "-", 2)[0], 10, 32)
		if err != nil {
			return 0
		}
		return uint32(u)
	}
	return u.u
}

type AmbrData struct {
	Response int `json:"response"`
	Data     struct {
		Items map[string]struct {
			ID   Uint32orString `json:"id"`
			Name string         `json:"name"`
			Icon string         `json:"icon"`
		} `json:"items"`
	} `json:"data"`
}

func (s *Service) initData(url string, m map[uint32]*Data) error {
	body, err := s.cacheGet(url)
	if err != nil {
		return err
	}
	var data AmbrData
	if err = json.Unmarshal(body, &data); err != nil {
		return err
	}
	for _, item := range data.Data.Items {
		m[item.ID.Uint32()] = &Data{ID: item.ID.Uint32(), Name: item.Name, Icon: item.Icon}
	}
	return nil
}
