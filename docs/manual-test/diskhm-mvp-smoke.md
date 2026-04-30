# DiskHM MVP Smoke Checklist

## Build

- [ ] From `web/`, run `npm run build`.
- [ ] Confirm `internal/webassets/dist/` contains `index.html` and built assets.
- [ ] Run `GOPROXY=https://goproxy.cn,direct "C:\Program Files\Go\bin\go.exe" test -mod=readonly ./...`.
- [ ] Run `GOPROXY=https://goproxy.cn,direct "C:\Program Files\Go\bin\go.exe" build -o diskhmd ./cmd/diskhmd`.

## Runtime

- [ ] Start `diskhmd` with a local config file.
- [ ] Open `http://127.0.0.1:9789/` and confirm the frontend loads.
- [ ] Request `http://127.0.0.1:9789/api/health` and confirm it returns `{"status":"ok"}`.
- [ ] Confirm a direct asset URL under `/assets/` is served.

## Installer

- [ ] Review `packaging/systemd/diskhm.service` for the expected `ExecStart`.
- [ ] Run `scripts/install-local.sh` on a Linux systemd host with `diskhmd` already built.
- [ ] Confirm `/usr/local/bin/diskhm`, `/etc/diskhm/config.yaml`, and `/var/lib/diskhm` exist after install.
- [ ] Confirm `systemctl status diskhm.service` shows the unit enabled and started.
