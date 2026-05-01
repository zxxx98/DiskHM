# DiskHM MVP Smoke Checklist

## Build

- [ ] From `web/`, run `npm run build`.
- [ ] Confirm `internal/webassets/dist/` contains `index.html` and built assets.
- [ ] Run `GOPROXY=https://goproxy.cn,direct "C:\Program Files\Go\bin\go.exe" test -mod=readonly ./...`.
- [ ] Run `GOPROXY=https://goproxy.cn,direct "C:\Program Files\Go\bin\go.exe" build -o diskhmd ./cmd/diskhmd`.

## Runtime

- [ ] Start `diskhmd daemon --config /etc/diskhm/config.yaml`.
- [ ] Open `http://127.0.0.1:9789/` and confirm the frontend loads.
- [ ] Sign in with the configured token and confirm the app redirects to `/disks`.
- [ ] Confirm `/disks` shows discovered disk rows on a populated Linux host.
- [ ] Confirm `/topology` shows live route content rather than the scaffold placeholder text.
- [ ] Confirm `/settings` shows the configured quiet grace seconds.
- [ ] Confirm `/events` shows an empty-state or live event entries rather than scaffold text.
- [ ] Request `http://127.0.0.1:9789/api/health` and confirm it returns `{"status":"ok"}`.
- [ ] Confirm a direct asset URL under `/assets/` is served.
- [ ] Trigger `Sleep now` on a supported HDD and confirm a new event appears in `/events`.
- [ ] Trigger `Refresh (wake disk)` on a supported disk and confirm a new event appears in `/events`.

## Installer

- [ ] Review `packaging/systemd/diskhm.service` for the expected `ExecStart`.
- [ ] Run `scripts/install-local.sh` on a Linux systemd host with `diskhmd` already built.
- [ ] Confirm `/usr/local/bin/diskhm`, `/etc/diskhm/config.yaml`, and `/var/lib/diskhm` exist after install.
- [ ] Confirm `systemctl status diskhm.service` shows the unit enabled and started.
- [ ] Confirm `sudo diskhm start|stop|enable|disable|uninstall` work as expected.
