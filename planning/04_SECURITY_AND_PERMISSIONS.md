# SkipTUI: Security, Permissions & Threat Model

## 1. Privilege & Capabilities Model

Configuring Linux network interfaces, namespaces, and OpenVPN/WireGuard TUN devices natively requires administrative capabilities. SkipTUI is designed to minimize privilege escalation and supports three tiers of operation.

```
+--------------------------------------------------------------------------------+
|                             PRIVILEGE TIERS                                    |
+--------------------------+---------------------------+-------------------------+
| Tier 1: Linux Caps       | Tier 2: Sudo / Elevated   | Tier 3: Universal Fallback
| (`sudo make setcap`)     | (`sudo ./bin/skiptui`)    | (EnvProxy Engine)       |
+--------------------------+---------------------------+-------------------------+
| - Binary granted         | - Runs with sudo          | - 100% unprivileged     |
|   `cap_net_admin,`       | - Full system namespace   | - Environment proxy     |
|   `cap_sys_admin+ep`     |   orchestration           |   routing for SOCKS/HTTP|
| - `openvpn` granted      |                           |   (no TUN device needed)|
|   `cap_net_admin+ep`     |                           |                         |
| - Fast native `netns`    |                           |                         |
|   without full root      |                           |                         |
+--------------------------+---------------------------+-------------------------+
```

### 1.1 Recommended Setup: Linux Capabilities
Instead of running SkipTUI entirely as root, users grant network admin capabilities to the binary and OpenVPN helper:
```bash
sudo make setcap
```
This executes:
```bash
sudo setcap 'cap_net_admin,cap_sys_admin+ep' ./bin/skiptui
sudo setcap 'cap_net_admin+ep' /usr/sbin/openvpn
```
This enables SkipTUI and OpenVPN to create network namespaces, allocate TUN devices (`ioctl TUNSETIFF`), and set isolated routing tables without granting full unrestricted root filesystem access.

---

## 2. Threat Analysis & Leak Vectors

### 2.1 Threat Matrix & Mitigations
| Threat / Leak Vector | Attack / Vulnerability Scenario | SkipTUI Mitigation Strategy |
| :--- | :--- | :--- |
| **DNS Leak** | Process ignores namespace DNS and queries host `127.0.0.53` (systemd-resolved). | Private mount namespace with custom `/etc/resolv.conf`; sanitize `DBUS_*` environment variables to prevent IPC escape. |
| **WebRTC Local IP Leak** | Browser gathers host LAN IP (`192.168.1.X`) via STUN requests. | Process inside namespace only sees `10.0.0.2` and loopback `127.0.0.1`. Host interfaces (`eth0`/`wlan0`) are invisible. |
| **Local Subnet Escape** | Compromised process tries to connect to local devices (e.g. `192.168.1.1` router admin). | Isolated routing table has zero routes to host subnet; default route goes exclusively to virtual TUN device. |
| **IPv6 Leak** | App connects over IPv6 directly via host ISP because proxy is IPv4-only. | Strict disablement of IPv6 inside namespace (`sysctl net.ipv6.conf.all.disable_ipv6=1`) unless explicitly configured in profile. |
| **Tunnel Crash Leak** | Proxy worker process crashes, leaving process to fall back to host network. | Fail-closed routing: without the TUN worker, traffic has no path and drops immediately (hardware kill-switch effect). |

---

## 3. Credential & Key Security

SkipTUI manages sensitive credentials including WireGuard Private Keys, Shadowsocks passwords, and HTTP/SOCKS5/OpenVPN credentials.

### 3.1 Storage Security Principles
1. **Strict File Permissions**: All configuration files (`~/.config/skiptui/config.yaml` and `~/.config/skiptui/profiles/*_auth.txt`) are created with `0600` permissions (read/write only by the current user).
2. **Auto-Linked Auth Files**: OpenVPN `auth-user-pass` credentials are automatically created and bound to the corresponding `.ovpn` file.
3. **Environment Variable Injection**: Credentials can be referenced via environment variables (e.g. `SOCKS5_PASSWORD_ENV`) rather than stored in plain text.
4. **Memory Zeroing**: Sensitive keys in memory are scrubbed when profiles are unloaded.

---

## 4. Signal Trapping & Namespace Garbage Collection

Abrupt exits (e.g. `SIGINT`, terminal window closure, crash) could leave orphan network namespaces or stale virtual interfaces in the kernel.

### 4.1 Lifecycle Guarantees
1. **Signal Trapping**: SkipTUI registers handlers for `SIGINT`, `SIGTERM`, `SIGHUP`, and `SIGQUIT`.
2. **Namespace Tagging**: All namespaces created by SkipTUI follow the naming pattern: `skiptui-<timestamp>-<random_id>`.
3. **Startup Orphan Sweeper**: On launch, SkipTUI scans for any lingering `skiptui-*` namespaces that do not correspond to active PIDs and destroys them safely.
