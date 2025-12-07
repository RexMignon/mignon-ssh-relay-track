# SSH Relay Track User Guide(English/简体中文)

For **internal network tunneling** to function correctly, you must **modify your **remote SSH server's** configuration file.

1. Log in to your remote SSH server.

2. Edit the SSH configuration file with root privileges:

```

sudo vim /etc/ssh/sshd_config

```

3. Ensure the file contains the following **two lines** of configuration, and they are not commented out with `#`:

```shell

# This is the master switch for all port forwarding functions; it must be yes

AllowTcpForwarding yes

# This allows the reverse tunnel port to be accessed from the public network. The yes option means that regardless of whether 127.0.0.1 is specified, it will tunnel to the server's 0.0.0.0

# The meaning of clientspecified is to let the user choose, rather than forcibly exposing to 0.0.0.0

GatewayPorts clientspecified

```

4. After saving the file, **restart the SSH service** for the configuration to take effect:

``shell

sudo systemctl restart For sshd (Debian) and other distributions, please refer to the documentation.

```
And, enable the firewall on the corresponding port.

How to prevent your password or username from being stolen?

You can create a user with minimal privileges:

```shell
sudo useradd -r -M -s /sbin/nologin -c "SSH Tunnel User for Intranet Penetration" tunneluser

# Then set the password
sudo passwd tunneluser

```
`sudo useradd tunneluser`: This is the core of the command, creating a user named `tunneluser`.

**`-r` (`--system`)**: Creates a **system account**. System accounts are typically used to run specific services or processes, not for live logins. Their UID (User ID) is usually assigned in a low range (e.g., 100-999) to distinguish them from regular users (usually starting from 1000).

` ...11111 **`-M` (`--no-create-home`):** Prevents the creation of a home directory for this user (`/home/tunneluser`). Since this user is only used to establish tunnels and doesn't need to store any files, there's no need to create a home directory. This also reduces system clutter and potential attack surfaces.

**`-s /sbin/nologin` (`--shell`):** This is the most critical security setting. It sets the user's default shell to `/sbin/nologin`. This means that when `tunneluser` attempts to log in interactively via SSH, the system immediately rejects the session and disconnects the connection, instead of providing a command-line interface like `bash` or `sh`. The user cannot execute commands like `ls`, `cd`, `rm`, etc.

**`-c "..." (`--comment`):** Adds a descriptive comment for the user. The comment "SSH Tunnel User for Intranet Penetration" clearly explains the intent behind creating this user. `sudo passwd tunneluser`: Sets a password for the newly created user. Although this user cannot log in to the shell, the SSH service can still use this password to authenticate their identity and authorize them to establish a tunnel.


# ssh relay Track 使用指南

为了让**内网穿透功能能够正常工作，您必须**修改您的**远程SSH服务器**的配置文件。

1. 登录到您的远程SSH服务器。

2. 使用 `root` 权限编辑 SSH 配置文件：

   ```
   sudo vim /etc/ssh/sshd_config
   ```

3. 确保文件中有以下**两行**配置，并且它们没有被 `#` 注释掉：

   ```shell
   # 这是所有端口转发功能的总开关，必须为 yes
   AllowTcpForwarding yes
   
   # 这允许反向隧道的端口被公网访问，yes 选项是无论是指定127.0.0.1都会穿透到服务器的0.0.0.0
   # 而clientspecified 的含义是让用户自己选择, 而不是强制暴露到0.0.0.0
   GatewayPorts clientspecified
   ```

4. 保存文件后，**重启SSH服务**以使配置生效：

   ```shell
   sudo systemctl restart sshd(Debian)其他发行版请另行查阅
   ```

并且 , 开启对应端口的防火墙

如何不会被盗取密码或者用户名?

可以创建一个最小权限的用户:

```shell
sudo useradd -r -M -s /sbin/nologin -c "SSH Tunnel User for Intranet Penetration" tunneluser
# 然后设置密码
sudo passwd tunneluser
```

`sudo useradd tunneluser`: 这是命令的核心，创建一个名为 `tunneluser` 的用户。

**`-r` (`--system`)**: 创建一个**系统账户**。系统账户通常用于运行特定的服务或进程，而不是给真人登录使用。它们的 UID（用户ID）通常会分配在一个较低的范围（例如 100-999），以便和普通用户（通常从 1000 开始）区分开。

**`-M` (`--no-create-home`)**: **不为该用户创建家目录** (`/home/tunneluser`)。因为这个用户只是用来建立隧道，它不需要存储任何文件，所以没有必要为其创建家目录。这也能减少系统的混乱和潜在的攻击面。

**`-s /sbin/nologin` (`--shell`)**: 这是**最关键的安全设置**。它将用户的默认 Shell 设置为 `/sbin/nologin`。这意味着当 `tunneluser` 尝试通过 SSH 进行交互式登录时，系统会立即拒绝会话并断开连接，而不是提供一个像 `bash` 或 `sh` 这样的命令行界面。用户无法执行 `ls`, `cd`, `rm` 等任何命令。

**`-c "..."` (`--comment`)**: 为用户添加一段描述性注释。这里的注释 "SSH Tunnel User for Intranet Penetration"（用于内网穿透的SSH隧道用户）非常清晰地说明了创建这个用户的意图。

`sudo passwd tunneluser`: 为这个刚刚创建的用户设置密码。虽然该用户无法登录 Shell，但 SSH 服务仍然可以使用这个密码来验证其身份，以便授权其建立隧道。

