# DiskHM

DiskHM is a Linux disk sleep manager with an embedded Go API and frontend bundle.

## Current default

DiskHM now listens on:

- `0.0.0.0:9789`

That means both `localhost` and the host's LAN/VPN IP can access the web UI by default.

## Quick install

You can install DiskHM on a Linux host with one command:

```bash
curl -fsSL https://raw.githubusercontent.com/zxxx98/DiskHM/main/install.sh | sudo bash
```

The one-line installer:

- downloads the current `main` branch source archive
- builds `diskhmd` with the local Go toolchain
- runs the local install flow
- creates `/etc/diskhm/config.yaml`
- installs and starts `diskhm.service`

Requirements for the one-line installer:

- Linux
- `systemd`
- `curl`
- `tar`
- `go`

## Web access

By default the service binds to `0.0.0.0:9789`.

You can open the UI from:

- the same machine: [http://127.0.0.1:9789](http://127.0.0.1:9789)
- another machine on the same network: `http://<host-ip>:9789`

Health check:

```bash
curl http://127.0.0.1:9789/api/health
```

Expected response:

```json
{"status":"ok"}
```

## Installed commands

After installation, the binary is installed as:

- `/usr/local/bin/diskhm`

Daemon mode:

```bash
diskhm daemon --config /etc/diskhm/config.yaml
```

Service management:

```bash
sudo diskhm start
sudo diskhm stop
sudo diskhm enable
sudo diskhm disable
sudo diskhm uninstall
```

Command behavior:

- `start`: start `diskhm.service`
- `stop`: stop `diskhm.service`
- `enable`: enable `diskhm.service` on boot
- `disable`: disable `diskhm.service` on boot
- `uninstall`: stop and disable the service, then remove:
  - `/usr/local/bin/diskhm`
  - `/etc/systemd/system/diskhm.service`
  - `/etc/diskhm`
  - `/var/lib/diskhm`

## Configuration

DiskHM reads its config from:

- `/etc/diskhm/config.yaml`

Default generated config:

```yaml
server:
  listen_addr: 0.0.0.0:9789
security:
  token_hash: bootstrap-token-hash-change-me
sleep:
  quiet_grace_seconds: 10
```

If you change the config:

```bash
sudo systemctl restart diskhm.service
```

## Local development

Build the frontend bundle:

```bash
cd web
npm install
npm run build
```

Run the Go test suite:

```bash
GOPROXY=https://goproxy.cn,direct "C:\Program Files\Go\bin\go.exe" test -mod=readonly ./...
```

Build the daemon binary:

```bash
GOPROXY=https://goproxy.cn,direct "C:\Program Files\Go\bin\go.exe" build -o diskhmd ./cmd/diskhmd
```

Run the daemon directly:

```bash
./diskhmd daemon --config /etc/diskhm/config.yaml
```

Install locally on a Linux systemd host:

```bash
sudo ./scripts/install-local.sh
```
