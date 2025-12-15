package ssh_penetrate

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	log "mignon-ssh-port-forworder-dev/app/pkg/logging"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

// StartReverseSSHTunnel 启动反向隧道
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

			proxyUrl, _ := getProxyURLFromEnv()
			proxyMsg := "直连"
			if proxyUrl != nil {
				proxyMsg = fmt.Sprintf("代理:%s", proxyUrl.Host)
			}
			log.Logger.Info(fmt.Sprintf("[RevTunnel-Manager] 正在连接 SSH [%s] (尝试: %d/%d)...", proxyMsg, retryCount+1, maxRetries))

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
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	// --- 修改开始: 使用代理拨号 ---
	var client *ssh.Client
	var err error

	proxyDialer := getEnvDialer()

	// 1. 建立底层 TCP 连接
	conn, err := proxyDialer.Dial("tcp", sshAddr)
	if err != nil {
		return fmt.Errorf("拨号失败(检查代理): %w", err)
	}

	// 2. 建立 SSH 连接
	c, chans, reqs, err := ssh.NewClientConn(conn, sshAddr, config)
	if err != nil {
		err := conn.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("SSH 握手失败: %w", err)
	}
	client = ssh.NewClient(c, chans, reqs)
	// --- 修改结束 ---

	defer func(client *ssh.Client) {
		err := client.Close()
		if err != nil {
			return
		}
	}(client)

	// 3. 请求远程服务器监听端口
	remoteListener, err := client.Listen("tcp", remoteListenAddr)
	if err != nil {
		return fmt.Errorf("请求远程监听失败 (端口可能被占用): %w", err)
	}
	defer func(remoteListener net.Listener) {
		err := remoteListener.Close()
		if err != nil {
			return
		}
	}(remoteListener)

	log.Logger.Info(fmt.Sprintf("[RevTunnel-Session] 映射建立: 远程[%s] -> 本地[%s]", remoteListenAddr, localTargetAddr))

	sessionErrChan := make(chan error, 1)

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

	go func() {
		for {
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

func handleReverseForwarding(remoteConn net.Conn, localTargetAddr string) {
	defer func(remoteConn net.Conn) {
		err := remoteConn.Close()
		if err != nil {

		}
	}(remoteConn)

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

	copyConn := func(dst, src net.Conn, result chan<- error) {
		_, err := io.Copy(dst, src)
		result <- err
	}

	resCh := make(chan error, 2)
	go copyConn(localConn, remoteConn, resCh)
	go copyConn(remoteConn, localConn, resCh)

	<-resCh
}

// --- 辅助函数：复制一份在这里，保持包的独立性 ---

func getEnvDialer() proxy.Dialer {
	u, err := getProxyURLFromEnv()
	if err != nil || u == nil {
		return proxy.Direct
	}
	// 创建代理 Dialer
	dialer, err := proxy.FromURL(u, proxy.Direct)
	if err != nil {
		log.Logger.Warn(fmt.Sprintf("[Proxy] 代理初始化失败: %v, 使用直连", err))
		return proxy.Direct
	}
	return dialer
}

func getProxyURLFromEnv() (*url.URL, error) {
	keys := []string{
		"ALL_PROXY", "all_proxy",
		"HTTPS_PROXY", "https_proxy",
		"HTTP_PROXY", "http_proxy",
	}

	var proxyStr string
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			proxyStr = v
			break
		}
	}

	if proxyStr == "" {
		return nil, nil
	}

	if !strings.Contains(proxyStr, "://") {
		proxyStr = "socks5://" + proxyStr
	}

	return url.Parse(proxyStr)
}
