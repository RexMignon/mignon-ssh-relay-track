

# SSH Relay Track User Guide(English/简体中文)

Para que el **túnel de red interna** funcione correctamente, debe **modificar el archivo de configuración de su **servidor SSH remoto**.

1. Inicie sesión en su servidor SSH remoto.

2. Edite el archivo de configuración de SSH con privilegios de root:

```

sudo vim /etc/ssh/sshd_config

```

3. Asegúrese de que el archivo contenga las siguientes **dos líneas** de configuración y que no estén comentadas con `#`:

```shell

# This is the master switch for all port forwarding functions; it must be yes

AllowTcpForwarding yes

# This allows the reverse tunnel port to be accessed from the public network. The yes option means that regardless of whether 127.0.0.1 is specified, it will tunnel to the server's 0.0.0.0

# The meaning of clientspecified is to let the user choose, rather than forcibly exposing to 0.0.0.0

GatewayPorts clientspecified

```

4. Después de guardar el archivo, **reinicie el servicio SSH** para que la configuración surta efecto:

```shell

sudo systemctl restart For sshd (Debian) and other distributions, please refer to the documentation.

```
Además, active el firewall en el puerto correspondiente.

¿Cómo evitar que su contraseña o nombre de usuario sean robados?

Puede crear un usuario con privilegios mínimos:

```shell
sudo useradd -r -M -s /sbin/nologin -c "SSH Tunnel User for Intranet Penetration" tunneluser

# Then set the password
sudo passwd tunneluser

```
`sudo useradd tunneluser`: Este es el núcleo del comando, que crea un usuario llamado `tunneluser`.

**`-r` (`--system`)**: Crea una **cuenta de sistema**. Las cuentas de sistema suelen utilizarse para ejecutar servicios o procesos específicos, no para inicios de sesión interactivos. Su UID (ID de usuario) generalmente se asigna en un rango bajo (por ejemplo, 100-999) para distinguirlas de los usuarios regulares (que suelen comenzar desde 1000).

` ...11111 **`-M` (`--no-create-home`):** Evita la creación de un directorio principal para este usuario (`/home/tunneluser`). Dado que este usuario solo se utiliza para establecer túneles y no necesita almacenar archivos, no es necesario crear un directorio principal. Esto también reduce el desorden del sistema y las superficies de ataque potenciales.

**`-s /sbin/nologin` (`--shell`):** Esta es la configuración de seguridad más crítica. Establece la shell predeterminada del usuario en `/sbin/nologin`. Esto significa que cuando `tunneluser` intente iniciar sesión de forma interactiva a través de SSH, el sistema rechazará inmediatamente la sesión y desconectará la conexión, en lugar de proporcionar una interfaz de línea de comandos como `bash` o `sh`. El usuario no podrá ejecutar comandos como `ls`, `cd`, `rm`, etc.

**`-c "..." (`--comment`):** Añade un comentario descriptivo para el usuario. El comentario "SSH Tunnel User for Intranet Penetration" explica claramente la intención detrás de la creación de este usuario. `sudo passwd tunneluser`: Establece una contraseña para el usuario recién creado. Aunque este usuario no puede iniciar sesión en la shell, el servicio SSH todavía puede utilizar esta contraseña para autenticar su identidad y autorizarlo para establecer un túnel.


# ssh relay Track 使用指南

Para que la **función de túnel de red interna** funcione correctamente, debe **modificar el archivo de configuración de su **servidor SSH remoto**.

1. Inicie sesión en su servidor SSH remoto.

2. Edite el archivo de configuración de SSH con permisos de `root`:

   ```
   sudo vim /etc/ssh/sshd_config
   ```

3. Asegúrese de que el archivo contenga las siguientes **dos líneas** de configuración y que no estén comentadas con `#`:

   ```shell
   # 这是所有端口转发功能的总开关，必须为 yes
   AllowTcpForwarding yes
   
   # 这允许反向隧道的端口被公网访问，yes 选项是无论是指定127.0.0.1都会穿透到服务器的0.0.0.0
   # 而clientspecified 的含义是让用户自己选择, 而不是强制暴露到0.0.0.0
   GatewayPorts clientspecified
   ```

4. Después de guardar el archivo, **reinicie el servicio SSH** para que la configuración surta efecto:

   ```shell
   sudo systemctl restart sshd(Debian)其他发行版请另行查阅
   ```

Además, active el firewall en el puerto correspondiente.

¿Cómo evitar que su contraseña o nombre de usuario sean robados?

Puede crear un usuario con privilegios mínimos:

```shell
sudo useradd -r -M -s /sbin/nologin -c "SSH Tunnel User for Intranet Penetration" tunneluser
# 然后设置密码
sudo passwd tunneluser
```

`sudo useradd tunneluser`: Este es el núcleo del comando, que crea un usuario llamado `tunneluser`.

**`-r` (`--system`)**: Crea una **cuenta de sistema**. Las cuentas de sistema suelen utilizarse para ejecutar servicios o procesos específicos, no para inicios de sesión interactivos. Su UID (ID de usuario) generalmente se asigna en un rango bajo (por ejemplo, 100-999) para distinguirlas de los usuarios regulares (que suelen comenzar desde 1000).

**`-M` (`--no-create-home`)**: **No crea un directorio principal para este usuario** (`/home/tunneluser`). Dado que este usuario solo se utiliza para establecer túneles y no necesita almacenar archivos, no es necesario crear un directorio principal. Esto también reduce el desorden del sistema y las superficies de ataque potenciales.

**`-s /sbin/nologin` (`--shell`)**: Esta es la **configuración de seguridad más crítica**. Establece la shell predeterminada del usuario en `/sbin/nologin`. Esto significa que cuando `tunneluser` intente iniciar sesión de forma interactiva a través de SSH, el sistema rechazará inmediatamente la sesión y desconectará la conexión, en lugar de proporcionar una interfaz de línea de comandos como `bash` o `sh`. El usuario no podrá ejecutar comandos como `ls`, `cd`, `rm`, etc.

**`-c "..."` (`--comment`)**: Añade un comentario descriptivo para el usuario. El comentario "SSH Tunnel User for Intranet Penetration" (Usuario de túnel SSH para penetración de red interna) explica claramente la intención detrás de la creación de este usuario.

`sudo passwd tunneluser`: Establece una contraseña para el usuario recién creado. Aunque este usuario no puede iniciar sesión en la shell, el servicio SSH todavía puede utilizar esta contraseña para autenticar su identidad y autorizarlo para establecer un túnel.
