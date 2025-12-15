package ssh_forward

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

// StartSSHTunnel 启动 SSH 隧道
func StartSSHTunnel(sshAddr, user, password, localAddr, remoteAddr string) (func(), <-chan error) {
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
				log.Logger.Info("[Tunnel-Manager] 用户主动停止隧道")
				return
			default:
			}

			// 尝试获取代理地址用于日志打印
			proxyUrl, _ := getProxyURLFromEnv()
			proxyMsg := "直连"
			if proxyUrl != nil {
				proxyMsg = fmt.Sprintf("代理:%s", proxyUrl.Host)
			}
			log.Logger.Warn(fmt.Sprintf("[Tunnel-Manager] 尝试建立连接 [%s] (尝试次数: %d/%d)...", proxyMsg, retryCount+1, maxRetries))

			err := runTunnelSession(sshAddr, user, password, localAddr, remoteAddr, stopCtxChan)

			if err == nil {
				return
			}

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

func runTunnelSession(sshAddr, user, password, localAddr, remoteAddr string, stopSignal <-chan struct{}) error {
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

	// 1. 获取支持代理的 Dialer
	proxyDialer := getEnvDialer()

	// 2. 建立底层 TCP 连接
	conn, err := proxyDialer.Dial("tcp", sshAddr)
	if err != nil {
		return fmt.Errorf("拨号失败(检查代理设置): %w", err)
	}

	// 3. 建立 SSH 连接
	c, chans, reqs, err := ssh.NewClientConn(conn, sshAddr, config)
	if err != nil {
		err := conn.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("SSH 握手失败: %w", err)
	}
	client = ssh.NewClient(c, chans, reqs)

	defer func(client *ssh.Client) {
		err := client.Close()
		if err != nil {
			return
		}
	}(client)

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return err
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			return
		}
	}(listener)

	log.Logger.Info(fmt.Sprintf("[Tunnel-Session] 隧道建立: %s -> %s -> %s", localAddr, sshAddr, remoteAddr))

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
					case sessionErrChan <- fmt.Errorf("心跳失败: %w", err):
					default:
					}
					return
				}
			}
		}
	}()

	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				select {
				case <-stopSignal:
				default:
					select {
					case sessionErrChan <- fmt.Errorf("监听器 Accept 错误: %w", err):
					default:
					}
				}
				return
			}
			go handleForwarding(client, localConn, remoteAddr)
		}
	}()

	select {
	case <-stopSignal:
		return nil
	case err := <-sessionErrChan:
		return err
	}
}

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

	copyConn := func(dst, src net.Conn, result chan<- error) {
		_, err := io.Copy(dst, src)
		result <- err
	}

	resCh := make(chan error, 2)
	go copyConn(remoteConn, localConn, resCh)
	go copyConn(localConn, remoteConn, resCh)
	<-resCh
}

// --- 辅助函数：从环境变量获取代理 ---

// getEnvDialer 返回一个代理 Dialer 或者直连 Dialer
func getEnvDialer() proxy.Dialer {
	u, err := getProxyURLFromEnv()
	if err != nil || u == nil {
		return proxy.Direct
	}

	// 强制把 dialer 当作 SOCKS5 处理 (x/net/proxy 主要支持 socks5)
	// 如果用户填的是 http://127.0.0.1:7890，我们这里也尝试用 SOCKS5 协议去连这个端口
	// 因为 SSH 是 TCP 协议，不能走 HTTP 代理
	dialer, err := proxy.FromURL(u, proxy.Direct)
	if err != nil {
		log.Logger.Warn(fmt.Sprintf("[Proxy] 代理地址解析失败: %v, 将使用直连", err))
		return proxy.Direct
	}
	return dialer
}

// getProxyURLFromEnv 按照优先级读取环境变量
func getProxyURLFromEnv() (*url.URL, error) {
	// 优先级: ALL_PROXY > HTTPS_PROXY > HTTP_PROXY (及其小写)
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

	// 容错处理：如果用户没写协议头 (e.g. "127.0.0.1:7890")，默认补上 socks5://
	if !strings.Contains(proxyStr, "://") {
		proxyStr = "socks5://" + proxyStr
	}

	u, err := url.Parse(proxyStr)
	if err != nil {
		return nil, err
	}
	return u, nil
}
