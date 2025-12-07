package manager

import (
	"fmt"
	"mignon-ssh-port-forworder-dev/app/cmd/ssh_basic/ssh_forward"
	"mignon-ssh-port-forworder-dev/app/cmd/ssh_basic/ssh_penetrate"
	"mignon-ssh-port-forworder-dev/app/pkg/config"
	log "mignon-ssh-port-forworder-dev/app/pkg/logging"
	"sync"
)

// TunnelEvent 用于通知 UI 或日志层发生了什么
type TunnelEvent struct {
	ID         string // 隧道的唯一标识 (ServerId_LinkId)
	LinkName   string // 隧道的易读名称
	Error      string // [修改] 改为 string 类型，确保前端能正确显示错误文本
	IsStopped  bool   // true 表示收到停止信号
	ServerName string
}

// TunnelManager 管理所有隧道生命周期
type TunnelManager struct {
	// 存储所有活跃隧道的停止函数: map[TunnelID]StopFunc
	activeTunnels map[string]func()

	// 存储活跃隧道的“配置指纹”，用于比对关键参数是否变化
	// map[TunnelID] "User@Host:Port|Local->Remote"
	activeSignatures map[string]string

	mu sync.RWMutex

	// 全局事件通道
	EventChan chan TunnelEvent
}

var (
	Instance = NewTunnelManager()
)

func NewTunnelManager() *TunnelManager {
	return &TunnelManager{
		activeTunnels:    make(map[string]func()),
		activeSignatures: make(map[string]string),
		EventChan:        make(chan TunnelEvent, 100),
	}
}

// Sync 智能同步: 仅在配置的关键参数(IP,端口,密码等)发生变化时才重启隧道
func (tm *TunnelManager) Sync(cfg *config.IConfig) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	visitedIDs := make(map[string]bool)

	for _, serverGroup := range cfg.Config {
		if !serverGroup.IsOpen {
			continue
		}

		for _, link := range serverGroup.LinkGroup {
			if !link.IsOpen {
				continue
			}

			tunnelID := generateID(serverGroup.Id, link.Id)
			visitedIDs[tunnelID] = true

			newSig := computeConfigSignature(serverGroup, link)

			stopFunc, exists := tm.activeTunnels[tunnelID]
			oldSig := tm.activeSignatures[tunnelID]

			if !exists {
				// 情况 A: 新隧道 -> 启动
				log.Logger.Info(fmt.Sprintf("[Manager] 新增隧道，正在启动: %s", link.Name))
				tm.startTunnelUnsafe(tunnelID, serverGroup, link, newSig)
			} else if oldSig != newSig {
				// 情况 B: 参数变更 -> 重启
				log.Logger.Info(fmt.Sprintf("[Manager] 关键配置变更，正在重启隧道: %s", link.Name))
				stopFunc()
				delete(tm.activeTunnels, tunnelID)
				delete(tm.activeSignatures, tunnelID)
				tm.startTunnelUnsafe(tunnelID, serverGroup, link, newSig)
			}
		}
	}

	// 清理不需要的隧道
	for id, stopFunc := range tm.activeTunnels {
		if !visitedIDs[id] {
			log.Logger.Error(fmt.Sprintf("[Manager] 配置已移除或关闭，停止隧道: %s", id))
			stopFunc()
			delete(tm.activeTunnels, id)
			delete(tm.activeSignatures, id)
		}
	}
}

// StopAll 停止所有
func (tm *TunnelManager) StopAll() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for id, stopFunc := range tm.activeTunnels {
		stopFunc()
		log.Logger.Info(fmt.Sprintf("[Manager] 停止隧道: %s", id))
	}
	tm.activeTunnels = make(map[string]func())
	tm.activeSignatures = make(map[string]string)
}

// GetRunningIDs 获取所有运行中的 ID
func (tm *TunnelManager) GetRunningIDs() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	ids := make([]string, 0, len(tm.activeTunnels))
	for id := range tm.activeTunnels {
		ids = append(ids, id)
	}
	return ids
}

func generateID(serverId, linkId string) string {
	return fmt.Sprintf("%s_%s", serverId, linkId)
}

func computeConfigSignature(server config.IConfigGroup, link config.IConfigLinkGroup) string {
	return fmt.Sprintf("%v|%s:%s@%s:%d|%s:%d->%s:%d",
		link.IsPenetrate,
		server.Username, server.Password, server.ServerHost, server.ServerPort,
		link.LocalHost, link.LocalPort, link.RemoteHost, link.RemotePort,
	)
}

// startTunnelUnsafe 内部启动逻辑
func (tm *TunnelManager) startTunnelUnsafe(id string, server config.IConfigGroup, link config.IConfigLinkGroup, signature string) {
	sshAddr := fmt.Sprintf("%s:%d", server.ServerHost, server.ServerPort)

	var stopFunc func()
	var errChan <-chan error

	if link.IsPenetrate {
		remoteListen := fmt.Sprintf("%s:%d", link.RemoteHost, link.RemotePort)
		localTarget := fmt.Sprintf("%s:%d", link.LocalHost, link.LocalPort)
		stopFunc, errChan = ssh_penetrate.StartReverseSSHTunnel(sshAddr, server.Username, server.Password, remoteListen, localTarget)
	} else {
		localListen := fmt.Sprintf("%s:%d", link.LocalHost, link.LocalPort)
		remoteTarget := fmt.Sprintf("%s:%d", link.RemoteHost, link.RemotePort)
		stopFunc, errChan = ssh_forward.StartSSHTunnel(sshAddr, server.Username, server.Password, localListen, remoteTarget)
	}

	tm.activeTunnels[id] = stopFunc
	tm.activeSignatures[id] = signature

	go func() {
		err, ok := <-errChan
		if ok {
			log.Logger.Error(fmt.Sprintf("[Manager] 隧道 [%s] 异常退出: %v", link.Name, err))
			tm.mu.Lock()
			if tm.activeSignatures[id] == signature {
				delete(tm.activeTunnels, id)
				delete(tm.activeSignatures, id)
			}
			tm.mu.Unlock()

			// 发送事件
			tm.EventChan <- TunnelEvent{
				ServerName: server.ServerName,
				ID:         id,
				LinkName:   link.Name,
				Error:      err.Error(),
			}
		}
	}()
}
