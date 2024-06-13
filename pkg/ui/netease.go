package ui

import (
	"encoding/json"
	"go-musicfox/pkg/lastfm"
	"strconv"
	"time"

	tea "github.com/anhoder/bubbletea"
	"github.com/anhoder/netease-music/service"
	"github.com/anhoder/netease-music/util"
	"github.com/mattn/go-runewidth"
	"github.com/telanflow/cookiejar"
	"go-musicfox/pkg/configs"
	"go-musicfox/pkg/constants"
	"go-musicfox/pkg/storage"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
)

type NeteaseModel struct {
	WindowWidth    int
	WindowHeight   int
	isListeningKey bool
	program        *tea.Program
	user           *structs.User
	lastfm         *lastfm.Client
	lastfmUser     *storage.LastfmUser

	// startup
	startup *StartupModel

	// main ui
	*MainUIModel

	// login
	loginModel *LoginModel

	// search
	searchModel *SearchModel
}

// NewNeteaseModel get netease model
func NewNeteaseModel(loadingDuration time.Duration) (m *NeteaseModel) {
	m = new(NeteaseModel)
	m.isListeningKey = !configs.ConfigRegistry.StartupShow
	m.lastfm = lastfm.NewClient()

	// startup
	m.startup = NewStartup()
	m.startup.TotalDuration = loadingDuration

	// main menu
	m.MainUIModel = NewMainUIModel(m)

	// login
	m.loginModel = NewLogin()

	// search
	m.searchModel = NewSearch()

	// 东亚
	runewidth.EastAsianWidth = false

	return
}

func (m *NeteaseModel) Init() tea.Cmd {
	projectDir := utils.GetLocalDataDir()

	// 全局文件Jar
	cookieJar, _ := cookiejar.NewFileJar(projectDir+"/cookie", nil)
	util.SetGlobalCookieJar(cookieJar)

	// DBManager初始化
	storage.DBManager = new(storage.LocalDBManager)

	// 获取用户信息
	go func() {
		defer utils.Recover(false)

		table := storage.NewTable()

		// 获取用户信息
		if jsonStr, err := table.GetByKVModel(storage.User{}); err == nil {
			if user, err := structs.NewUserFromLocalJson(jsonStr); err == nil {
				m.user = &user
			}
		}
		m.refreshMenuTitle() // 刷新界面用户名

		// 获取lastfm用户信息
		var lastfmUser storage.LastfmUser
		if jsonStr, err := table.GetByKVModel(&lastfmUser); err == nil {
			if err = json.Unmarshal(jsonStr, &lastfmUser); err == nil {
				m.lastfmUser = &lastfmUser
				m.lastfm.SetSession(lastfmUser.SessionKey)
			}
		}
		m.refreshMenuList()

		// 获取播放模式
		if jsonStr, err := table.GetByKVModel(storage.PlayMode{}); err == nil && len(jsonStr) > 0 {
			var playMode PlayMode
			if err = json.Unmarshal(jsonStr, &playMode); err == nil {
				m.player.mode = playMode
			}
		}

		// 获取音量
		if jsonStr, err := table.GetByKVModel(storage.Volume{}); err == nil && len(jsonStr) > 0 {
			var volume int
			if err = json.Unmarshal(jsonStr, &volume); err == nil {
				v, ok := m.player.Player.(storage.VolumeStorable)
				if ok {
					v.SetVolume(volume)
				}
			}
		}

		// 获取播放歌曲信息
		if jsonStr, err := table.GetByKVModel(storage.PlayerSnapshot{}); err == nil && len(jsonStr) > 0 {
			var snapshot storage.PlayerSnapshot
			if err = json.Unmarshal(jsonStr, &snapshot); err == nil {
				player := m.player
				player.curSongIndex = snapshot.CurSongIndex
				player.playlist = snapshot.Playlist
				player.playlistUpdateAt = snapshot.PlaylistUpdateAt
				player.curSong = player.playlist[player.curSongIndex]
				player.playingMenuKey = "from_local_db" // 启动后，重置菜单Key，避免很多问题
			}
		}
		m.Rerender()

		// 签到
		if configs.ConfigRegistry.StartupSignIn {
			var lastSignIn int
			if jsonStr, err := table.GetByKVModel(storage.LastSignIn{}); err == nil && len(jsonStr) > 0 {
				_ = json.Unmarshal(jsonStr, &lastSignIn)
			}
			today, err := strconv.Atoi(time.Now().Format("20060102"))
			if m.user != nil && err == nil && lastSignIn != today {
				// 手机签到
				signInService := service.DailySigninService{}
				signInService.Type = "0"
				signInService.DailySignin()
				// PC签到
				signInService.Type = "1"
				signInService.DailySignin()
				// 云贝签到
				ybSignService := service.YunbeiSigninService{}
				ybSignService.Signin()

				_ = table.SetByKVModel(storage.LastSignIn{}, today)

				utils.Notify(utils.NotifyContent{
					Title: "签到成功",
					Text:  "今日手机、PC端签到成功",
					Url:   constants.AppGithubUrl,
				})
			}
		}

		// 检查更新
		if configs.ConfigRegistry.StartupCheckUpdate && utils.CheckUpdate() {
			utils.Notify(utils.NotifyContent{
				Title:  "发现新版本",
				Text:   "点击去看看呗",
				Url:    constants.AppLatestReleases,
				Sender: "default",
			})
		}
	}()

	if configs.ConfigRegistry.StartupShow {
		return tickStartup(time.Nanosecond)
	}

	return tickMainUI(time.Nanosecond)
}

func (m *NeteaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	switch msgWithType := msg.(type) {
	case tea.KeyMsg:
		k := msgWithType.String()
		// 登录界面输入q不退出
		if m.pageType == PtMain && (k == "q" || k == "Q" || k == "ctrl+c") {
			m.startup.quitting = true
			m.Close()
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.WindowHeight = msgWithType.Height
		m.WindowWidth = msgWithType.Width
	}

	// Hand off the message and model to the approprate update function for the
	// appropriate view based on the current state.
	if configs.ConfigRegistry.StartupShow && !m.startup.loaded {
		if _, ok := msg.(tea.WindowSizeMsg); ok {
			m.MainUIModel.update(msg, m)
		}
		return m.startup.update(msg, m)
	}

	switch m.pageType {
	case PtLogin:
		return updateLogin(msg, m)
	case PtSearch:
		return m.searchModel.update(msg, m)
	}

	return m.MainUIModel.update(msg, m)
}

func (m *NeteaseModel) View() string {
	if m.startup.quitting || m.WindowWidth <= 0 || m.WindowHeight <= 0 {
		return ""
	}

	if configs.ConfigRegistry.StartupShow && !m.startup.loaded {
		return m.startup.view(m)
	}

	switch m.pageType {
	case PtLogin:
		return loginView(m)
	case PtSearch:
		return m.searchModel.view(m)
	}

	return m.MainUIModel.view(m)
}

func (m *NeteaseModel) BindProgram(program *tea.Program) {
	m.program = program
}

func (m *NeteaseModel) Rerender() {
	if m.program != nil {
		m.program.Rerender(m.View())
	}
}

func (m *NeteaseModel) Close() {
	m.MainUIModel.Close()
}
