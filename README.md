# DiskHM

DiskHM is a Linux disk sleep manager with an embedded Go API and frontend bundle.

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

Install locally on a systemd host:

```bash
sudo ./scripts/install-local.sh
```
