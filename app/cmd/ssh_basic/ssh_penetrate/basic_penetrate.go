package ssh_penetrate

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	log "mignon-ssh-port-forworder-dev/app/pkg/logging"

	"golang.org/x/crypto/ssh"
)

// StartReverseSSHTunnel 启动反向隧道 (内网穿透)
// 相当于: ssh -NR remoteListenAddr:localTargetAddr user@sshAddr
//
// 参数说明:
// sshAddr:        SSH 服务器地址 (e.g., "1.2.3.4:22")
// user, password: SSH 认证信息
// remoteListenAddr: 远程服务器上监听的地址 (e.g., "0.0.0.0:8080" 或 "127.0.0.1:8080")
// localTargetAddr:  本地需要暴露的服务地址 (e.g., "127.0.0.1:3000")
func StartReverseSSHTunnel(sshAddr, user, password, remoteListenAddr, localTargetAddr string) (func(), <-chan error) {
	errChan := make(chan error, 1)
	stopCtxChan := make(chan struct{})
	var once sync.Once

	stopFunc := func() {
		once.Do(func() {
			close(stopCtxChan)
		})
	}

	go func() {
		const maxRetries = 5
		retryCount := 0

		for {
			select {
			case <-stopCtxChan:
				log.Logger.Info("[RevTunnel-Manager] 用户主动停止隧道")
				return
			default:
			}

			log.Logger.Info(fmt.Sprintf("[RevTunnel-Manager] 正在连接 SSH 并请求远程监听 (尝试: %d/%d)...", retryCount+1, maxRetries))

			// 启动反向隧道会话
			err := runReverseSession(sshAddr, user, password, remoteListenAddr, localTargetAddr, stopCtxChan)

			if err == nil {
				return
			}

			log.Logger.Warn(fmt.Sprintf("[RevTunnel-Manager] 连接断开: %v", err))
			retryCount++

			if retryCount >= maxRetries {
				errMsg := fmt.Errorf("反向隧道重连失败 (%d次): %v", maxRetries, err)
				log.Logger.Error(fmt.Sprintf("%v", errMsg))
				select {
				case errChan <- errMsg:
				default:
				}
				return
			}

			log.Logger.Warn("[RevTunnel-Manager] 5秒后重连...")
			select {
			case <-time.After(5 * time.Second):
			case <-stopCtxChan:
				return
			}
		}
	}()

	return stopFunc, errChan
}

func runReverseSession(sshAddr, user, password, remoteListenAddr, localTargetAddr string, stopSignal <-chan struct{}) error {
	// 1. SSH 配置
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	// 2. 连接 SSH 服务器
	client, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		return err
	}
	defer func(client *ssh.Client) {
		err := client.Close()
		if err != nil {

		}
	}(client)

	// 3. 请求远程服务器监听端口 (这是 -R 的核心)
	// 注意: 如果 SSHD 配置 GatewayPorts no，这里绑定 0.0.0.0 可能只会生效在 127.0.0.1
	remoteListener, err := client.Listen("tcp", remoteListenAddr)
	if err != nil {
		return fmt.Errorf("请求远程监听失败 (端口可能被占用): %w", err)
	}
	defer func(remoteListener net.Listener) {
		err := remoteListener.Close()
		if err != nil {

		}
	}(remoteListener)

	log.Logger.Info(fmt.Sprintf("[RevTunnel-Session] 映射建立: 远程[%s] -> 本地[%s]", remoteListenAddr, localTargetAddr))

	sessionErrChan := make(chan error, 1)

	// 协程 A: 心跳保活 (和 -L 模式一样重要)
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
					case sessionErrChan <- fmt.Errorf("SSH 心跳失败: %w", err):
					default:
					}
					return
				}
			}
		}
	}()

	// 协程 B: 处理来自远程服务器的连接请求
	go func() {
		for {
			// 这里 Accept 的是远程服务器上转过来的连接
			remoteConn, err := remoteListener.Accept()
			if err != nil {
				select {
				case <-stopSignal:
				default:
					select {
					case sessionErrChan <- fmt.Errorf("远程监听器 Accept 错误: %w", err):
					default:
					}
				}
				return
			}
			// 处理转发
			go handleReverseForwarding(remoteConn, localTargetAddr)
		}
	}()

	select {
	case <-stopSignal:
		return nil
	case err := <-sessionErrChan:
		return err
	}
}

// handleReverseForwarding 将远程来的流量转发给本地服务
func handleReverseForwarding(remoteConn net.Conn, localTargetAddr string) {
	defer func(remoteConn net.Conn) {
		err := remoteConn.Close()
		if err != nil {

		}
	}(remoteConn)

	// 拨号本地真实服务 (例如 localhost:8080)
	localConn, err := net.Dial("tcp", localTargetAddr)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("[RevForward] 连接本地目标失败 [%s]: %v", localTargetAddr, err))
		return
	}
	defer func(localConn net.Conn) {
		err := localConn.Close()
		if err != nil {

		}
	}(localConn)

	// 双向拷贝
	copyConn := func(dst, src net.Conn, result chan<- error) {
		_, err := io.Copy(dst, src)
		result <- err
	}

	resCh := make(chan error, 2)
	go copyConn(localConn, remoteConn, resCh) // 远程 -> 本地
	go copyConn(remoteConn, localConn, resCh) // 本地 -> 远程

	<-resCh
}
