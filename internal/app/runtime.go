package app

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/example/diskhm/internal/config"
	"github.com/example/diskhm/internal/discovery"
	"github.com/example/diskhm/internal/domain"
	"github.com/example/diskhm/internal/refresh"
	"github.com/example/diskhm/internal/scheduler"
)

type EventStore interface {
	AppendEvent(context.Context, domain.Event) error
	ListEvents(context.Context, int) ([]domain.Event, error)
	UpsertDisk(context.Context, domain.Disk) error
}

type DiskTask struct {
	DiskID      string
	Kind        string
	State       string
	RequestedAt time.Time
	ExecuteAt   time.Time
	LastError   string
}

type Runtime struct {
	config   config.Config
	discover discovery.Service
	events   EventStore
	refresh  *refresh.Service
	sleep    *scheduler.Service

	mu          sync.RWMutex
	snapshot    domain.DiscoverySnapshot
	refreshedAt time.Time
	tasks       map[string]DiskTask
	cancels     map[string]context.CancelFunc

	subMu            sync.RWMutex
	subscribers      map[int]chan domain.Event
	nextSubscriberID int
}

func NewRuntime(cfg config.Config, discover discovery.Service, events EventStore) *Runtime {
	return &Runtime{
		config:      cfg,
		discover:    discover,
		events:      events,
		tasks:       make(map[string]DiskTask),
		cancels:     make(map[string]context.CancelFunc),
		subscribers: make(map[int]chan domain.Event),
	}
}

func (r *Runtime) SetRefreshService(service refresh.Service) {
	r.refresh = &service
}

func (r *Runtime) SetSchedulerService(service scheduler.Service) {
	r.sleep = &service
}

func (r *Runtime) Snapshot(ctx context.Context) (domain.DiscoverySnapshot, error) {
	r.mu.RLock()
	if !r.refreshedAt.IsZero() {
		snapshot := r.snapshot
		r.mu.RUnlock()
		return snapshot, nil
	}
	r.mu.RUnlock()

	return r.refreshSnapshot(ctx)
}

func (r *Runtime) Topology(ctx context.Context) (domain.TopologyGraph, error) {
	snapshot, err := r.Snapshot(ctx)
	if err != nil {
		return domain.TopologyGraph{}, err
	}

	nodes := make([]domain.TopologyNode, 0, len(snapshot.Disks)+len(snapshot.Mounts))
	edges := make([]domain.TopologyEdge, 0, len(snapshot.Mounts))

	for _, disk := range snapshot.Disks {
		nodes = append(nodes, domain.TopologyNode{
			ID:    disk.ID,
			Kind:  "disk",
			Label: disk.Path,
		})
	}

	for _, mount := range snapshot.Mounts {
		mountID := "mount-" + mount.Target
		nodes = append(nodes, domain.TopologyNode{
			ID:    mountID,
			Kind:  "mount",
			Label: mount.Target,
		})
		if mount.DiskID != "" {
			edges = append(edges, domain.TopologyEdge{
				From: mount.DiskID,
				To:   mountID,
			})
		}
	}

	return domain.TopologyGraph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

func (r *Runtime) ListDisks(ctx context.Context) ([]domain.DiskView, error) {
	snapshot, err := r.Snapshot(ctx)
	if err != nil {
		return nil, err
	}

	r.mu.RLock()
	refreshedAt := r.refreshedAt
	tasks := make(map[string]DiskTask, len(r.tasks))
	for id, task := range r.tasks {
		tasks[id] = task
	}
	r.mu.RUnlock()

	mountsByDiskID := make(map[string][]string)
	for _, mount := range snapshot.Mounts {
		if mount.DiskID == "" {
			continue
		}
		mountsByDiskID[mount.DiskID] = append(mountsByDiskID[mount.DiskID], mount.Target)
	}

	freshness := "cached"
	if !refreshedAt.IsZero() && time.Since(refreshedAt) < 5*time.Second {
		freshness = "live"
	}

	disks := make([]domain.DiskView, 0, len(snapshot.Disks))
	for _, disk := range snapshot.Disks {
		view := domain.DiskView{
			ID:               disk.ID,
			Name:             disk.Name,
			Path:             disk.Path,
			Model:            disk.Model,
			PowerState:       "unknown",
			RefreshFreshness: freshness,
			Unsupported:      !disk.Rotational,
			Mounts:           mountsByDiskID[disk.ID],
		}
		if task, ok := tasks[disk.ID]; ok {
			view.Task = &domain.DiskTaskView{
				Kind:      task.Kind,
				State:     task.State,
				ExecuteAt: task.ExecuteAt,
				LastError: task.LastError,
			}
		}
		disks = append(disks, view)
	}

	return disks, nil
}

func (r *Runtime) Settings(context.Context) (domain.SettingsView, error) {
	return domain.SettingsView{
		QuietGraceSeconds: r.config.Sleep.QuietGraceSeconds,
	}, nil
}

func (r *Runtime) ListEvents(ctx context.Context, limit int) ([]domain.Event, error) {
	if r.events == nil {
		return []domain.Event{}, nil
	}
	return r.events.ListEvents(ctx, limit)
}

func (r *Runtime) SubscribeEvents() (<-chan domain.Event, func()) {
	r.subMu.Lock()
	id := r.nextSubscriberID
	r.nextSubscriberID++
	ch := make(chan domain.Event, 16)
	r.subscribers[id] = ch
	r.subMu.Unlock()

	unsubscribe := func() {
		r.subMu.Lock()
		if existing, ok := r.subscribers[id]; ok {
			delete(r.subscribers, id)
			close(existing)
		}
		r.subMu.Unlock()
	}

	return ch, unsubscribe
}

func (r *Runtime) SleepNow(ctx context.Context, diskID string) error {
	disk, err := r.diskByID(ctx, diskID)
	if err != nil {
		return err
	}
	if !disk.Rotational {
		return domain.ErrUnsupportedDevice
	}
	if r.sleep == nil {
		return errors.New("scheduler service is not configured")
	}

	runCtx, cancel := context.WithCancel(context.Background())
	if err := r.setTask(diskID, DiskTask{
		DiskID:      diskID,
		Kind:        "sleep_now",
		State:       "scheduled",
		RequestedAt: time.Now(),
	}, cancel); err != nil {
		cancel()
		return err
	}

	if err := r.recordEvent(ctx, domain.Event{DiskID: diskID, Kind: "sleep_now_requested", Message: "sleep-now requested"}); err != nil {
		return err
	}

	go r.runSleepNow(runCtx, disk)
	return nil
}

func (r *Runtime) SleepAfter(ctx context.Context, diskID string, minutes int) error {
	disk, err := r.diskByID(ctx, diskID)
	if err != nil {
		return err
	}
	if !disk.Rotational {
		return domain.ErrUnsupportedDevice
	}
	if r.sleep == nil {
		return errors.New("scheduler service is not configured")
	}

	executeAt := time.Now().Add(time.Duration(minutes) * time.Minute)

	runCtx, cancel := context.WithCancel(context.Background())
	if err := r.setTask(diskID, DiskTask{
		DiskID:      diskID,
		Kind:        "sleep_after",
		State:       "scheduled",
		RequestedAt: time.Now(),
		ExecuteAt:   executeAt,
	}, cancel); err != nil {
		cancel()
		return err
	}

	if err := r.recordEvent(ctx, domain.Event{DiskID: diskID, Kind: "sleep_after_requested", Message: "sleep-after requested"}); err != nil {
		return err
	}

	go func() {
		timer := time.NewTimer(time.Until(executeAt))
		defer timer.Stop()
		select {
		case <-runCtx.Done():
			return
		case <-timer.C:
			r.runSleepNow(runCtx, disk)
		}
	}()

	return nil
}

func (r *Runtime) CancelSleep(ctx context.Context, diskID string) error {
	if _, err := r.diskByID(ctx, diskID); err != nil {
		return err
	}

	r.mu.Lock()
	cancel, ok := r.cancels[diskID]
	if ok {
		delete(r.cancels, diskID)
	}
	r.tasks[diskID] = DiskTask{
		DiskID:      diskID,
		Kind:        "sleep",
		State:       "canceled",
		RequestedAt: time.Now(),
	}
	r.mu.Unlock()

	if ok {
		cancel()
	}

	return r.recordEvent(ctx, domain.Event{DiskID: diskID, Kind: "sleep_canceled", Message: "sleep task canceled"})
}

func (r *Runtime) RefreshSafe(ctx context.Context, diskID string) error {
	if _, err := r.diskByID(ctx, diskID); err != nil {
		return err
	}
	if r.refresh != nil {
		disk, err := r.diskByID(ctx, diskID)
		if err != nil {
			return err
		}
		if err := r.refresh.SafeRefresh(ctx, disk); err != nil {
			return err
		}
	}
	if _, err := r.refreshSnapshot(ctx); err != nil {
		return err
	}
	return r.recordEvent(ctx, domain.Event{DiskID: diskID, Kind: "refresh_safe", Message: "safe refresh requested"})
}

func (r *Runtime) RefreshWake(ctx context.Context, diskID string) error {
	disk, err := r.diskByID(ctx, diskID)
	if err != nil {
		return err
	}
	if r.refresh != nil {
		if err := r.refresh.WakeRefresh(ctx, disk); err != nil {
			return err
		}
	}
	if _, err := r.refreshSnapshot(ctx); err != nil {
		return err
	}
	return r.recordEvent(ctx, domain.Event{DiskID: diskID, Kind: "refresh_wake", Message: "wake refresh requested"})
}

func (r *Runtime) diskByID(ctx context.Context, diskID string) (domain.Disk, error) {
	snapshot, err := r.Snapshot(ctx)
	if err != nil {
		return domain.Disk{}, err
	}
	for _, disk := range snapshot.Disks {
		if disk.ID == diskID {
			return disk, nil
		}
	}
	return domain.Disk{}, domain.ErrDiskNotFound
}

func (r *Runtime) refreshSnapshot(ctx context.Context) (domain.DiscoverySnapshot, error) {
	snapshot, err := r.discover.Snapshot(ctx)
	if err != nil {
		return domain.DiscoverySnapshot{}, err
	}

	if r.events != nil {
		for _, disk := range snapshot.Disks {
			if err := r.events.UpsertDisk(ctx, disk); err != nil {
				return domain.DiscoverySnapshot{}, err
			}
		}
	}

	r.mu.Lock()
	r.snapshot = snapshot
	r.refreshedAt = time.Now()
	r.mu.Unlock()

	return snapshot, nil
}

func (r *Runtime) setTask(diskID string, task DiskTask, cancel context.CancelFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.tasks[diskID]; ok && existing.State != "canceled" && existing.State != "failed" && existing.State != "sleeping" {
		return domain.ErrTaskConflict
	}

	r.tasks[diskID] = task
	r.cancels[diskID] = cancel
	return nil
}

func (r *Runtime) runSleepNow(ctx context.Context, disk domain.Disk) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		r.updateTaskState(disk.ID, "executing", "")
		err := r.sleep.RunSleepNow(ctx, disk)
		if err == nil {
			r.updateTaskState(disk.ID, "sleeping", "")
			_ = r.recordEvent(context.Background(), domain.Event{DiskID: disk.ID, Kind: "sleep_succeeded", Message: "disk entered standby"})
			r.clearCancel(disk.ID)
			return
		}

		if errors.Is(err, scheduler.ErrDiskBusy) {
			r.updateTaskState(disk.ID, "waiting_idle", err.Error())
			_ = r.recordEvent(context.Background(), domain.Event{DiskID: disk.ID, Kind: "sleep_waiting_idle", Message: "waiting for disk to become idle"})
			timer := time.NewTimer(time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				continue
			}
		}

		r.updateTaskState(disk.ID, "failed", err.Error())
		_ = r.recordEvent(context.Background(), domain.Event{DiskID: disk.ID, Kind: "sleep_failed", Message: err.Error()})
		r.clearCancel(disk.ID)
		return
	}
}

func (r *Runtime) updateTaskState(diskID string, state string, lastError string) {
	r.mu.Lock()
	task := r.tasks[diskID]
	task.State = state
	task.LastError = lastError
	r.tasks[diskID] = task
	r.mu.Unlock()
}

func (r *Runtime) clearCancel(diskID string) {
	r.mu.Lock()
	delete(r.cancels, diskID)
	r.mu.Unlock()
}

func (r *Runtime) recordEvent(ctx context.Context, event domain.Event) error {
	if r.events != nil {
		if err := r.events.AppendEvent(ctx, event); err != nil {
			return err
		}
	}

	r.subMu.RLock()
	defer r.subMu.RUnlock()
	for _, subscriber := range r.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}

	return nil
}
