package helper

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var slices = map[int][][]string{
	3: {{
		"https://act.hoyoverse.com/map_manage/20221122/4dccdca89bc40c396ee3adb3d2a54535_678692064171919141.png",
		"https://act.hoyoverse.com/map_manage/20221122/a733f5730bc89b2a8713c79217316a48_9138588374623790954.png",
		"https://act.hoyoverse.com/map_manage/20221122/4a780782734ba5aaf3e8f81a7ed9586c_290273923733944333.png",
		"https://act.hoyoverse.com/map_manage/20221122/ec01c15d9292f163e4e27cba4f9e6539_9212926219978415677.png",
	}, {
		"https://act.hoyoverse.com/map_manage/20221122/4a8df85626aab56bc3761b3468eda992_4957344763706963021.png",
		"https://act.hoyoverse.com/map_manage/20221122/b9a1465c2bf581f5ff51bf27fdc53f4f_178680297786250288.png",
		"https://act.hoyoverse.com/map_manage/20221122/9a7be9cca6eb748270c058e3deef4314_1630901094808151515.png",
		"https://act.hoyoverse.com/map_manage/20221122/b81e48c049406d1f0b690e0d89bcb219_4971424153465858852.png",
	}, {
		"https://act.hoyoverse.com/map_manage/20221122/7349aec11ad174fa6627efa309dc0f2f_637060848780817954.png",
		"https://act.hoyoverse.com/map_manage/20221122/59305ea0f07e522abb773a3ea6aca7f7_5483461357356348128.png",
		"https://act.hoyoverse.com/map_manage/20221122/ae348de45969306ac5fd30261a4d215a_1831881192516607234.png",
		"https://act.hoyoverse.com/map_manage/20221122/e7e07b6ee8002ebf84a0c90ce5a2d11a_6474304874717444837.png",
	}},
	5: {{
		"https://upload-static.hoyoverse.com/map_manage/20220329/d258137dc0e84fc8acbf77b7dc7115da_1941568151557226408.jpeg",
	}},
	7: {{
		"https://upload-static.hoyoverse.com/map_manage/20220329/c7ebe11c865d421541688319f57abb04_4008875850180854873.jpeg",
	}},
}

var zooms = map[int]float32{
	2: 0.0625,
	3: 0.125,
	4: 0.25,
	5: 0.5,
	6: 1.0,
}

func handleTile(w http.ResponseWriter, r *http.Request) {
	scene, _ := strconv.Atoi(r.URL.Query().Get("scene"))
	z, _ := strconv.Atoi(r.URL.Query().Get("z"))
	y, _ := strconv.Atoi(r.URL.Query().Get("y"))
	x, _ := strconv.Atoi(r.URL.Query().Get("x"))
	zz := 1 << (z - 2)
	i, j := y/zz, x/zz
	yy, xx := y%zz, x%zz
	switch scene {
	default:
		return
	case 3:
		if i < 0 || i > 2 || j < 0 || j > 3 || y < 0 || x < 0 {
			return
		}
	case 5, 7:
		if i != 0 || j != 0 || y < 0 || x < 0 {
			return
		}
	}
	url := slices[scene][i][j]
	url += fmt.Sprintf("?x-oss-process=image/resize,p_%v/crop,x_%d,y_%d,w_256,h_256/format,webp", zooms[z]*100, xx*256, yy*256)
	http.Redirect(w, r, url, http.StatusFound)
}

func handleIcon(w http.ResponseWriter, r *http.Request) {
	uri := strings.TrimPrefix(r.RequestURI, "/api/icon")
	http.Redirect(w, r, "https://api.ambr.top/assets"+uri, http.StatusFound)
}
