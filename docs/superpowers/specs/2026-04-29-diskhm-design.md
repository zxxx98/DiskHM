# DiskHM 设计文档

日期：2026-04-29

## 目标

DiskHM 是一个 Linux 原生磁盘休眠管理应用。它通过一个常驻 daemon 管理每块物理磁盘的休眠状态，提供“立即休眠”和“在 N 分钟之后休眠”两类操作，并且在真正执行休眠命令前等待磁盘 I/O 进入安静状态。

应用同时提供一个简单清爽的 Web 管理面板，用于查看磁盘状态、磁盘信息、休眠任务、事件日志和基础设置。面板必须避免为了展示更多信息而意外唤醒已经休眠的硬盘。

MVP 先支持普通 SATA/USB 机械硬盘，同时在架构和文档中保留 NAS 场景的后续开发方向，包括 mdadm、LVM、dm-crypt、ZFS 和 btrfs。

## MVP 范围

MVP 包含：

- Linux-only，使用 systemd 管理服务。
- 一个 Go 单二进制程序，内置 HTTP API 和 Web UI 静态资源。
- 每块物理磁盘独立管理。
- 支持“立即休眠”。
- 支持“在 N 分钟之后休眠”。
- 支持取消等待中的休眠任务。
- 休眠前等待无 I/O，而不是直接强制休眠。
- 展示磁盘基础信息、挂载点、I/O 状态、休眠状态、任务状态和信息新鲜度。
- 区分“安全刷新”和“可能唤醒硬盘的刷新”。
- 首页采用磁盘优先表格视图。
- 提供一个简单只读拓扑页，展示物理盘、分区、挂载点和已知逻辑关系。
- 提供事件日志和基础设置页。
- 提供一键安装脚本。

MVP 不包含：

- 自动识别并管理完整 RAID/ZFS/LVM 池级休眠策略。
- 文件系统卸载/重新挂载工作流。
- RAID rebuild、scrub、resilver 等维护任务协调。
- 多用户 RBAC。
- Windows/macOS 支持。
- 默认公网或局域网暴露。

## 产品原则

- 不为了补全 UI 信息而唤醒休眠硬盘。
- 可能唤醒硬盘的动作必须由用户明确触发。
- 默认展示缓存信息，并清楚标注缓存时间。
- 休眠操作进入任务状态，等待无 I/O 后再执行。
- UI 要适合日常运维使用：信息密度足够，但视觉上保持安静、清爽。
- 对不支持或不确定的硬件保持保守，不猜测执行危险命令。

## 技术架构

整体结构：

```text
systemd
  -> diskhmd
       -> discovery service
       -> I/O monitor
       -> sleep scheduler
       -> command executor
       -> metadata cache
       -> embedded HTTP API + Web UI
```

`diskhmd` 作为 root 服务运行，因为磁盘休眠命令需要权限。Web UI 不直接执行 shell 命令，只调用 daemon API。

默认监听：

- 地址：`127.0.0.1`
- 端口：`9789`

持久化路径：

- `/etc/diskhm/config.yaml`：服务配置、监听地址、认证策略、每盘偏好。
- `/var/lib/diskhm/diskhm.db`：SQLite 数据库，保存缓存、任务和事件。
- journald：默认日志位置。

实现选择：

- 后端：Go。
- 前端：React + TypeScript + Vite。
- UI 图标：lucide-react。
- 前端数据请求：TanStack Query。
- 数据库：SQLite。
- 静态资源：Go embed 打包进单二进制。
- 首版安装方式：安装脚本；deb/rpm 仓库放到后续阶段。

## 磁盘发现

磁盘发现优先使用低风险 Linux 元数据来源：

- `/sys/block/<dev>`：设备名、容量、rotational 标记、queue 属性、父子关系、I/O 统计。
- udev 数据库：model、serial、ID path、WWN、transport、稳定 by-id 名称。
- `/proc/self/mountinfo`：挂载点关系。
- sysfs holders/slaves：简单拓扑关系。
- `/proc/mdstat`：mdadm 提示信息，MVP 只读展示。

默认不对休眠盘执行以下可能读取块设备的操作：

- `blkid`
- `lsblk --fs`
- 直接读取 `/dev/<disk>` 或 `/dev/<partition>`
- 未经确认的 SMART/温度刷新

原因：当 udev 缓存缺失或信息不完整时，这类工具可能直接探测块设备，从而唤醒硬盘。

设备分类：

- 物理 HDD 是 MVP 的主要可控对象。
- SSD/NVMe 可展示，但默认禁用 HDD 风格休眠操作。
- `loop`、`ram`、`zram`、光驱和伪块设备默认隐藏。
- `md`、`dm-*`、LVM、dm-crypt、ZFS、btrfs 相关关系进入拓扑页，但 MVP 仍只对物理盘执行休眠。

## 信息刷新规则

Web UI 提供两类刷新：

### 安全刷新

默认刷新方式，不应唤醒休眠硬盘。

数据来源：

- sysfs
- udev 缓存
- mountinfo
- daemon 内存状态
- SQLite 缓存
- `/sys/block/<dev>/stat`

### 可能唤醒的刷新

需要用户显式确认。

包括：

- SMART 信息
- 温度
- 完整文件系统探测
- 任何可能读取块设备的命令
- 对设备类型不确定的 USB 桥接盘执行自动探测

SMART 策略：

- 优先使用 `smartctl -n standby`，避免在硬盘已经 standby 时继续查询。
- 不对休眠盘自动刷新 SMART。
- 成功读取的 SMART/温度写入缓存，并显示时间戳。
- 对 USB bridge、RAID controller 等不确定设备，要求用户配置设备类型或明确允许唤醒刷新。

电源状态策略：

- ATA/SATA 设备优先使用 `hdparm -C` 查询 standby 状态。
- 不可靠控制器返回的状态显示为 `unknown`。
- `unknown` 不等于 active，也不等于 sleeping；休眠操作仍必须经过 I/O 安静检测。

## I/O 安静检测

休眠任务必须通过 I/O 安静检测。

daemon 每秒采样 `/sys/block/<dev>/stat`。满足以下条件才认为磁盘安静：

- `in_flight == 0`
- 读写完成计数在安静窗口内不变化
- 安静窗口持续达到 `quiet_grace_seconds`

默认值：

- 采样间隔：1 秒。
- 安静窗口：10 秒。
- 最大等待时间：默认不限制，直到用户取消或任务失败。

挂载盘休眠前需要 flush：

- 优先对该磁盘相关挂载点执行 `syncfs`。
- 如果挂载点到物理盘的映射不明确，回退到系统级 `sync`。
- flush 会产生 I/O，因此 flush 完成后重新开始安静窗口计时。

## 休眠任务模型

每块磁盘同一时间最多一个休眠任务。

任务类型：

- `sleep_now`：立即进入休眠流程，但仍等待无 I/O。
- `sleep_after`：N 分钟后进入休眠流程，然后等待无 I/O。

任务状态：

- `scheduled`：等待目标时间。
- `flushing`：正在同步相关文件系统。
- `waiting_idle`：目标时间已到，但仍有 I/O。
- `executing`：正在执行休眠命令。
- `sleeping`：命令成功或状态确认已 standby。
- `failed`：执行失败。
- `canceled`：用户取消。

UI 必须显示等待原因，例如：

- 正在写入。
- 正在读取。
- 正在 flush。
- 设备不支持休眠命令。
- 设备状态未知。
- 休眠命令失败。

## 休眠命令

命令执行器按设备类型分派。

MVP：

- ATA/SATA HDD：使用 `hdparm -y /dev/<disk>` 进入 standby。
- SAT-capable USB HDD：在确认支持 ATA passthrough 时使用 ATA 路径。
- 不支持的 USB bridge：显示 unsupported，不执行命令。
- SSD/NVMe：默认只展示状态，不提供 HDD 休眠按钮。

后续：

- SCSI/SAS：评估 `sg_start --stop` 或 `sdparm`。
- USB bridge：建立兼容性 profile。
- NVMe：展示电源状态，不默认做 HDD 式休眠。
- 支持每盘命令 override，但需要清楚标注风险。

每次命令执行都写入事件：

- 设备。
- 命令类型。
- 开始时间和结束时间。
- exit status。
- stderr 摘要。
- 后续状态检测结果。

## HTTP API

API 初版：

- `GET /api/disks`：磁盘列表、缓存元数据、电源状态、I/O 状态、当前任务。
- `GET /api/disks/{id}`：单盘详情。
- `POST /api/disks/{id}/sleep-now`：创建立即休眠任务。
- `POST /api/disks/{id}/sleep-after`：创建延迟休眠任务，body 为 `{ "minutes": number }`。
- `POST /api/disks/{id}/cancel-sleep`：取消当前任务。
- `POST /api/disks/{id}/refresh-safe`：只刷新不会唤醒硬盘的数据。
- `POST /api/disks/{id}/refresh-wake`：执行可能唤醒硬盘的刷新。
- `GET /api/topology`：只读拓扑。
- `GET /api/events`：事件列表。
- `GET /api/events/stream`：SSE 实时事件流。
- `GET /api/settings`：读取设置。
- `PUT /api/settings`：更新设置。

错误响应格式：

```json
{
  "code": "disk_busy",
  "message": "Disk still has active I/O",
  "details": {
    "device": "sdc",
    "in_flight": 2
  }
}
```

## Web 面板设计

MVP 采用“磁盘优先表格 + 顶部状态汇总 + 简单拓扑标签页”。

### 首页

顶部状态汇总：

- 可控磁盘数。
- Active 磁盘数。
- Sleeping 磁盘数。
- Pending 任务数。
- Warning/unsupported 数。

磁盘表格列：

- 状态点。
- 磁盘标签：用户别名、`/dev/<disk>`、model。
- 挂载点/文件系统。
- 电源状态。
- 当前 I/O。
- 休眠策略或当前任务。
- 信息新鲜度。
- 操作。

行内操作：

- 立即休眠。
- N 分钟后休眠。
- 取消任务。
- 安全刷新。
- 可能唤醒的刷新。

展开行内容：

- by-id、serial、WWN。
- 容量、transport、rotational。
- 分区和挂载点。
- 上次 SMART/温度缓存。
- 最近事件。
- 单盘设置。

视觉风格：

- 管理后台风格，不做营销页。
- 中性背景、白色数据区域、紧凑间距。
- 状态用小圆点和 badge 表示。
- 主要按钮清晰，危险或可能唤醒操作必须二次确认。
- 移动端不追求完整表格体验，优先保证单盘卡片可用。

### 拓扑页

MVP 拓扑页只读，目标是为后续 NAS 支持铺路。

展示：

- 物理盘节点。
- 分区节点。
- 挂载点节点。
- 已发现的 mdadm/LVM/dm-crypt/ZFS/btrfs 关系。

限制：

- 不提供池级自动休眠。
- 不提供复杂依赖编辑。
- 不把拓扑发现失败当作休眠失败；物理盘控制仍可独立工作。

### 事件页

事件页展示：

- 休眠任务生命周期。
- 安全刷新和可能唤醒刷新。
- 命令失败。
- 不支持设备原因。
- daemon 启动、配置变更、设备增删。

### 设置页

设置页包含：

- 监听地址和端口。
- token 状态和轮换入口。
- 默认安静窗口。
- 每盘别名。
- 每盘命令支持 override。
- 是否展示 unsupported 设备。

## 安全模型

默认安全策略：

- 只监听 `127.0.0.1`。
- 首次安装生成随机 token。
- token hash 存储在 root 可读配置或状态文件中。
- 非 GET API 需要认证和 CSRF 防护。
- 休眠命令和可能唤醒刷新需要 UI 二次确认。

局域网访问：

- 必须显式配置。
- installer 在启用非 loopback bind 时打印警告。
- HTTPS 建议通过反向代理提供，MVP 不内置证书管理。

## 安装器

一键安装脚本职责：

- 检测 Linux、systemd、架构和必要命令。
- 安装 `/usr/local/bin/diskhm`。
- 创建 `/etc/diskhm/config.yaml`。
- 创建 `/var/lib/diskhm`。
- 创建并启用 `diskhm.service`。
- 生成初始 UI token。
- 输出本地访问地址和 token。

开发环境本地安装：

```bash
sudo ./scripts/install-local.sh
```

发布安装命令在 release 流程确定后写入 README，不在设计文档里固定占位 URL。

安装器不得直接覆盖已有配置；发现旧配置时先备份。

## 错误处理

核心错误类型：

- `unsupported_device`：设备没有安全可用的休眠命令。
- `disk_busy`：I/O 尚未安静。
- `flush_failed`：flush 失败。
- `command_failed`：休眠命令返回非零。
- `state_unknown`：无法安全判断电源状态。
- `refresh_requires_wake`：请求的数据需要可能唤醒的刷新。
- `permission_denied`：权限不足或认证失败。

UI 处理：

- 错误显示在对应磁盘行。
- 同步写入事件页。
- 对可恢复错误提供下一步操作，例如重试、取消、查看详情。
- 对 unsupported 不提供不可用按钮，只展示原因。

## 测试策略

单元测试：

- sysfs fixture 解析。
- mountinfo fixture 解析。
- udev metadata fixture 解析。
- 设备分类。
- I/O 安静判断。
- 休眠任务状态机。
- API 错误响应。

集成测试：

- 使用 fixture-backed discovery provider 启动 daemon。
- 覆盖 sleep-now、sleep-after、cancel 流程。
- 验证 safe refresh 不执行 wake-capable 命令。
- 验证 installer 在临时 root 下创建预期文件。

Linux 手工测试：

- SATA HDD + `hdparm`。
- SAT-capable USB HDD。
- 不支持 ATA passthrough 的 USB HDD。
- 已挂载且持续写入的磁盘。
- 已休眠磁盘执行 safe refresh。
- SMART refresh 在 standby 状态下跳过。
- 默认不监听局域网。

UI 测试：

- 首页表格渲染 active、sleeping、waiting_idle、unknown、unsupported、failed。
- 立即休眠进入 pending/waiting 状态，而不是直接显示成功。
- N 分钟后休眠显示倒计时，到点后进入 waiting_idle。
- 可能唤醒刷新必须确认。
- 窄屏下单盘操作可用且文字不溢出。

## 后续开发方向

### Phase 1：MVP

- Go daemon + React Web UI + systemd installer。
- sysfs/udev/mountinfo 发现。
- 每盘 sleep-now 和 sleep-after。
- I/O 安静检测。
- ATA `hdparm -y` 命令路径。
- safe refresh 和 SMART 缓存。
- 表格首页、简单拓扑页、事件页、设置页。

### Phase 2：NAS 关系增强

- 更完整的 mdadm、LVM、dm-crypt、ZFS、btrfs 拓扑模型。
- 池/卷组视图，说明逻辑卷由哪些物理盘组成。
- 在睡眠单盘可能影响活跃池时给出警告。
- 支持组操作，但最终仍分解到物理盘并通过 I/O gate。
- 更精确的 mount -> partition -> physical disk 映射，减少全局 `sync`。

### Phase 3：自动策略

- 每盘 idle-after 自动休眠。
- 维护窗口。
- 禁止休眠时间段。
- 基于进程、挂载点或路径的活动排除。
- 睡眠/唤醒历史统计。
- 休眠效果指标，例如节省时长和失败原因分布。

### Phase 4：硬件兼容性

- SCSI/SAS 命令支持。
- USB bridge 兼容性 profile。
- 每设备命令 override。
- 已知 USB/SATA bridge 的 SMART 安全 profile。
- 社区维护的硬件兼容性数据库。

### Phase 5：部署增强

- deb/rpm 包。
- 包仓库。
- 反向代理和 HTTPS 部署指南。
- 更完整的 token/session 管理。
- Prometheus metrics。
- 配置备份和恢复。

## 已定决策

- MVP 范围按“普通 SATA/USB HDD 优先，NAS 栈后续增强”执行。
- 首页采用磁盘优先表格。
- MVP 同时提供简单只读拓扑页。
- 后端使用 Go。
- 前端使用 React + TypeScript + Vite。
- 状态和事件存储使用 SQLite。
- 首版使用安装脚本，不先做 deb/rpm 仓库。

## 待确认但不阻塞 MVP

- 最终公开项目名是否继续使用 `DiskHM`。
- 首版发布渠道和下载域名。
- 是否需要官方 Docker 镜像。由于磁盘控制需要宿主机权限，Docker 不作为首推部署方式。

## 参考资料

- `lsblk` manual: https://man7.org/linux/man-pages/man8/lsblk.8.html
- `blkid` manual: https://man7.org/linux/man-pages/man8/blkid.8.html
- Linux block statistics documentation: https://www.kernel.org/doc/html/latest/block/stat.html
- `smartctl` manual: https://manpages.opensuse.org/Tumbleweed/smartmontools/smartctl.8.en.html
- `hdparm` manual: https://man7.org/linux/man-pages/man8/hdparm.8.html
