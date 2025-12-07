package main

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/energye/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	// 引入包
	"mignon-ssh-port-forworder-dev/app/cmd/ssh_basic/manager"
	"mignon-ssh-port-forworder-dev/app/pkg/config"
	"mignon-ssh-port-forworder-dev/app/pkg/logging"
)

// App struct
type App struct {
	ctx context.Context
	// errorCounts 用于记录每个隧道的连续错误次数 map[TunnelID]count
	errorCounts map[string]int
}

// 嵌入图标文件
//
//go:embed resources/rex.ico
var iconData []byte

// GetActiveTunnelIds 获取当前正在运行的隧道 ID 列表
func (a *App) GetActiveTunnelIds() []string {
	return manager.Instance.GetRunningIDs()
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		errorCounts: make(map[string]int),
	}
}

// Startup is called when the app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	logging.Logger.Info("[App] Startup: 正在同步配置并启动隧道服务...")

	// 1. 初始化 Manager (状态同步)
	manager.Instance.Sync(&config.SshConfig)

	// 2. 启动事件监听
	go a.monitorTunnelEvents()

	// 3. 启动系统托盘
	go systray.Run(a.onSystrayReady, a.onSystrayExit)
}

// onSystrayReady 系统托盘准备就绪时的回调
func (a *App) onSystrayReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("Mignon SSH Relay")
	systray.SetTooltip("Mignon SSH Relay Track")
	systray.SetOnClick(func(menu systray.IMenu) {
		a.showWindow()
	})

	systray.SetOnRClick(func(menu systray.IMenu) {
		err := menu.ShowMenu()
		if err != nil {
			return
		}
	})

	// 1. 显示窗口
	mShow := systray.AddMenuItem("显示窗口", "Show Main Window")
	mShow.Click(func() {
		a.showWindow()
	})

	systray.AddSeparator()

	// 2. 退出程序
	mQuit := systray.AddMenuItem("退出程序", "Quit Application")
	mQuit.Click(func() {
		systray.Quit()
		runtime.Quit(a.ctx)
	})
}

// onSystrayExit 系统托盘退出时的清理
func (a *App) onSystrayExit() {
	// 可以在这里做一些清理工作
}

// showWindow 辅助方法：显示并置顶窗口
func (a *App) showWindow() {
	runtime.WindowShow(a.ctx)
}

// monitorTunnelEvents 监听 Manager 的 channel 并转发给前端
// 这里增加了错误计数逻辑
func (a *App) monitorTunnelEvents() {
	for event := range manager.Instance.EventChan {
		// 发送 Wails 事件给前端更新 UI
		runtime.EventsEmit(a.ctx, "tunnel_event", event)

		if event.Error != "" {
			logging.Logger.Sugar().Errorf("[App-Event]服务器 %s 的隧道 %s 报错: %v", event.ServerName, event.LinkName, event.Error)

			// 异步弹窗，防止阻塞事件循环
			go func(e manager.TunnelEvent) {
				result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
					Type:          runtime.WarningDialog,
					Title:         "隧道连接警告",
					Message:       fmt.Sprintf("服务器%s 的隧道 [%s] 极不稳定，已连续失败超过 5 次。\n\n最新错误: %s\n\n请检查网络配置或服务器状态。", e.ServerName, e.LinkName, e.Error),
					DefaultButton: "知道了",
				})
				if err != nil {
					logging.Logger.Sugar().Error(err)
				}
				logging.Logger.Info(result)
			}(event)
		}

	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	logging.Logger.Sugar().Infof("Frontend Greet called with: %s", name)
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// ==========================================
// Config & Tunnel Control Methods
// ==========================================

// GetConfig 获取当前所有配置
func (a *App) GetConfig() config.IConfig {
	return config.SshConfig
}

// ForceReload 强制重新加载并同步所有隧道
func (a *App) ForceReload() {
	logging.Logger.Info("[App] ForceReload requested")
	manager.Instance.Sync(&config.SshConfig)
}

// ==========================================
// Server Group (服务器组) CRUD
// ==========================================

// AddServer 添加服务器组
func (a *App) AddServer(group config.IConfigGroup) {
	logging.Logger.Sugar().Infof("[App] 添加服务器组: %s", group.ServerName)
	config.SshConfig.AddIConfigGroup(&group)
	manager.Instance.Sync(&config.SshConfig)
}

// ModifyServer 修改服务器组
func (a *App) ModifyServer(id string, group config.IConfigGroup) {
	logging.Logger.Sugar().Infof("[App] 修改服务器组: %s", id)
	config.SshConfig.ModifyIConfigGroup(id, &group)
	manager.Instance.Sync(&config.SshConfig)
}

// DeleteServer 删除服务器组
func (a *App) DeleteServer(id string) {
	logging.Logger.Sugar().Infof("[App] 删除服务器组: %s", id)
	config.SshConfig.RemoveIConfigGroup(id)
	manager.Instance.Sync(&config.SshConfig)
}

// ==========================================
// Link (具体转发规则) CRUD
// ==========================================

// AddLink 添加转发规则
func (a *App) AddLink(serverId string, link config.IConfigLinkGroup) {
	logging.Logger.Sugar().Infof("[App] 添加 Link: %s -> Server: %s", link.Name, serverId)
	config.SshConfig.AddIConfigLinkGroup(serverId, link)
	manager.Instance.Sync(&config.SshConfig)
}

// ModifyLink 修改转发规则
func (a *App) ModifyLink(serverId, linkId string, link config.IConfigLinkGroup) {
	logging.Logger.Sugar().Infof("[App] 修改 Link: %s", linkId)
	config.SshConfig.ModifyIConfigLinkGroup(serverId, linkId, &link)
	manager.Instance.Sync(&config.SshConfig)
}

// DeleteLink 删除转发规则
func (a *App) DeleteLink(serverId string, linkId string) {
	logging.Logger.Sugar().Infof("[App] 删除 Link: %s", linkId)
	config.SshConfig.RemoveIConfigLinkGroup(serverId, linkId)
	manager.Instance.Sync(&config.SshConfig)
}

// ToggleLinkStatus 快速开关某个连接
func (a *App) ToggleLinkStatus(serverId string, linkId string, isOpen bool) {
	logging.Logger.Sugar().Infof("[App] 切换 Link 状态: %s -> %v", linkId, isOpen)
	serverIdx := -1
	for i, s := range config.SshConfig.Config {
		if s.Id == serverId {
			serverIdx = i
			break
		}
	}
	if serverIdx == -1 {
		return
	}

	linkIdx := -1
	for i, l := range config.SshConfig.Config[serverIdx].LinkGroup {
		if l.Id == linkId {
			linkIdx = i
			break
		}
	}
	if linkIdx == -1 {
		return
	}

	// 修改状态
	link := config.SshConfig.Config[serverIdx].LinkGroup[linkIdx]
	link.IsOpen = isOpen

	// 保存并同步
	config.SshConfig.ModifyIConfigLinkGroup(serverId, linkId, &link)
	manager.Instance.Sync(&config.SshConfig)
}

func (a *App) ThemeSwitch(switchDark bool) {
	config.SshConfig.IsDark = switchDark
	config.SshConfig.SetValue()
}

func (a *App) ModifyServers(serverId string, IsOpen bool) {
	serverIdx := -1
	for i, s := range config.SshConfig.Config {
		if s.Id == serverId {
			serverIdx = i
			break
		}
	}
	config.SshConfig.Config[serverIdx].IsOpen = IsOpen
	config.SshConfig.SetValue()
	manager.Instance.Sync(&config.SshConfig)
}

func (a *App) SetLanguage(isEnglish bool) {
	config.SshConfig.IsEnglish = isEnglish
	config.SshConfig.SetValue()
}
