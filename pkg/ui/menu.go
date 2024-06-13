package ui

type Hook func(m *NeteaseModel) bool

// IMenu menu interface
type IMenu interface {
    // IsPlayable 当前菜单是否可播放？
    IsPlayable() bool

    // ResetPlaylistWhenPlay 空格播放时，是否重置播放列表？
    ResetPlaylistWhenPlay() bool

    // GetMenuKey 菜单唯一Key
    GetMenuKey() string

    // MenuViews 获取子菜单View
    MenuViews() []MenuItem

    // MenuData 获取当前菜单的数据
    MenuData() interface{}

    // SubMenu 根据下标获取菜单Model
    SubMenu(model *NeteaseModel, index int) IMenu

    // BeforePrePageHook 切换上一页前的Hook
    BeforePrePageHook() Hook

    // BeforeNextPageHook 切换下一页前的Hook
    BeforeNextPageHook() Hook

    // BeforeEnterMenuHook 进入菜单项前的Hook
    BeforeEnterMenuHook() Hook

    // BeforeBackMenuHook 菜单返回前的Hook
    BeforeBackMenuHook() Hook

    // BottomOutHook 触底的Hook
    BottomOutHook() Hook

    // TopOutHook 触顶Hook
    TopOutHook() Hook
}

// IDjMenu dj menu interface
type IDjMenu interface {
	IMenu
}
