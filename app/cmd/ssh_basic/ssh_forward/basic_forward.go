package ssh_forward

import (
	"io"
	log "mignon-ssh-port-forworder-dev/app/pkg/logging"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

import (
	"fmt"
	"sync"
)

// StartSSHTunnel 启动 SSH 隧道
// 返回值:
// 1. stopFunc: 调用此函数以主动停止隧道
// 2. errorChan: 如果重连超过最大次数，错误会通过此通道发出来（此时隧道已停止）
func StartSSHTunnel(sshAddr, user, password, localAddr, remoteAddr string) (func(), <-chan error) {
	// 外部通知通道
	errChan := make(chan error, 1)
	// 内部控制通道
	stopCtxChan := make(chan struct{})
	var once sync.Once

	// 停止函数
	stopFunc := func() {
		once.Do(func() {
			close(stopCtxChan)
		})
	}

	// 守护协程
	go func() {
		const maxRetries = 5
		retryCount := 0

		// 外层循环：负责重连机制
		for {
			select {
			case <-stopCtxChan:
				log.Logger.Info("[Tunnel-Manager] 用户主动停止隧道")
				return
			default:
				// 继续执行
			}

			log.Logger.Warn(fmt.Sprintf("[Tunnel-Manager] 尝试建立连接 (尝试次数: %d/%d)...", retryCount+1, maxRetries))

			// 启动单次会话，该函数会阻塞直到连接断开或出错
			err := runTunnelSession(sshAddr, user, password, localAddr, remoteAddr, stopCtxChan)

			// 判断结果
			if err == nil {
				// 正常退出（说明收到了 stopCtxChan 信号）
				return
			}

			// 如果是运行中出错，进行重试逻辑
			log.Logger.Error(fmt.Sprintf("[Tunnel-Manager] 连接意外断开: %v", err))
			retryCount++

			if retryCount >= maxRetries {
				errMsg := fmt.Errorf("隧道重连失败达到上限 (%d次)，停止服务: %v", maxRetries, err)
				log.Logger.Error(fmt.Sprintf("%v", errMsg))
				select {
				case errChan <- errMsg:
				default:
				}
				return
			}

			// 等待一段时间后重连
			log.Logger.Info("[Tunnel-Manager] 3秒后尝试重连...")
			select {
			case <-time.After(3 * time.Second):
			case <-stopCtxChan:
				return
			}
		}
	}()
	return stopFunc, errChan
}

// runTunnelSession 负责一次完整的 SSH 连接生命周期
// 如果连接成功建立并保持，它会阻塞。
// 如果连接断开或建立失败，它会返回 error。
// 如果收到 stopSignal，它会返回 nil。
func runTunnelSession(sshAddr, user, password, localAddr, remoteAddr string, stopSignal <-chan struct{}) error {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	// 1. 建立 SSH 连接
	client, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		return err
	}
	defer func(client *ssh.Client) {
		err := client.Close()
		if err != nil {

		}
	}(client)

	// 2. 建立本地监听
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return err
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {

		}
	}(listener)

	log.Logger.Info(fmt.Sprintf("[Tunnel-Session] 隧道建立: %s -> %s -> %s", localAddr, sshAddr, remoteAddr))

	// 用于感知内部错误的通道
	sessionErrChan := make(chan error, 1)

	// 协程 A: 心跳检测
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stopSignal:
				return
			case <-ticker.C:
				_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
				if err != nil {
					select {
					case sessionErrChan <- fmt.Errorf("心跳失败: %w", err):
					default:
					}
					return
				}
			}
		}
	}()

	// 协程 B: 接收本地连接
	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				// 监听器关闭导致的错误，通常不需要作为 session 错误抛出，除非是异常关闭
				// 但为了简单，我们假设 Accept 报错就意味着 Listener 坏了
				select {
				case <-stopSignal:
					// 正常停止
				default:
					select {
					case sessionErrChan <- fmt.Errorf("监听器 Accept 错误: %w", err):
					default:
					}
				}
				return
			}
			// 处理单个连接转发，不阻塞 Accept 循环
			go handleForwarding(client, localConn, remoteAddr)
		}
	}()

	// 阻塞等待：要么用户停止，要么发生错误
	select {
	case <-stopSignal:
		return nil
	case err := <-sessionErrChan:
		return err
	}
}

// handleForwarding 具体的流转发逻辑 (保持你原有的核心逻辑，稍作资源清理优化)
func handleForwarding(sshClient *ssh.Client, localConn net.Conn, remoteAddr string) {
	defer func(localConn net.Conn) {
		err := localConn.Close()
		if err != nil {

		}
	}(localConn)

	remoteConn, err := sshClient.Dial("tcp", remoteAddr)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("[Forward] 远程拨号失败: %v", err))
		return
	}
	defer func(remoteConn net.Conn) {
		err := remoteConn.Close()
		if err != nil {

		}
	}(remoteConn)

	// 使用 WaitGroup 或 channel 确保双向 copy 完成
	copyConn := func(dst, src net.Conn, result chan<- error) {
		_, err := io.Copy(dst, src)
		result <- err
	}

	resCh := make(chan error, 2)
	go copyConn(remoteConn, localConn, resCh)
	go copyConn(localConn, remoteConn, resCh)
	<-resCh
}
