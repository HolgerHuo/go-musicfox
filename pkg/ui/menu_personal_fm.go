package ui

import (
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
	"time"
)

type PersonalFmMenu struct {
	menus []MenuItem
	songs []structs.Song
}

func NewPersonalFmMenu() *PersonalFmMenu {
	return new(PersonalFmMenu)
}

func (m *PersonalFmMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (m *PersonalFmMenu) IsPlayable() bool {
	return true
}

func (m *PersonalFmMenu) ResetPlaylistWhenPlay() bool {
	return true
}

func (m *PersonalFmMenu) GetMenuKey() string {
	return "personal_fm"
}

func (m *PersonalFmMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *PersonalFmMenu) MenuData() interface{} {
	return m.songs
}

func (m *PersonalFmMenu) SubMenu(*NeteaseModel, int) IMenu {
	return nil
}

func (m *PersonalFmMenu) BeforePrePageHook() Hook {
	// Nothing to do
	return nil
}

func (m *PersonalFmMenu) BeforeNextPageHook() Hook {
	// Nothing to do
	return nil
}

func (m *PersonalFmMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		// 已有数据
		if len(m.menus) > 0 && len(m.songs) > 0 {
			return true
		}

		personalFm := service.PersonalFmService{}
		code, response := personalFm.PersonalFm()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}

		// 响应中获取数据
		m.songs = utils.GetFmSongs(response)
		m.menus = GetViewFromSongs(m.songs)

		return true
	}
}

func (m *PersonalFmMenu) BottomOutHook() Hook {
	return func(model *NeteaseModel) bool {
		personalFm := service.PersonalFmService{}
		code, response := personalFm.PersonalFm()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}
		songs := utils.GetFmSongs(response)
		menus := GetViewFromSongs(songs)

		m.menus = append(m.menus, menus...)
		m.songs = append(m.songs, songs...)
		model.player.playlist = m.songs
		model.player.playlistUpdateAt = time.Now()

		return true
	}
}

func (m *PersonalFmMenu) TopOutHook() Hook {
	// Nothing to do
	return nil
}
