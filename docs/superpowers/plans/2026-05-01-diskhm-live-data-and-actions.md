# DiskHM Live Data And Actions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace placeholder API and UI behavior with live disk, topology, settings, and event data, and wire disk actions to real backend services.

**Architecture:** Add a small runtime service in `internal/app` that owns discovery snapshots, in-memory disk task state, SQLite-backed events, and action dispatch. Keep HTTP handlers thin by moving mapping and orchestration into that runtime, then switch the React routes to consume the live endpoints with TanStack Query and the existing SSE hook.

**Tech Stack:** Go, net/http, SQLite, React, TypeScript, TanStack Query, Vitest, Vite

---

### Task 1: Build Runtime State And Persistence Helpers

**Files:**
- Create: `internal/app/runtime.go`
- Create: `internal/app/runtime_test.go`
- Modify: `internal/store/sqlite/store.go`
- Modify: `internal/store/sqlite/store_test.go`
- Modify: `internal/domain/types.go`

- [ ] **Step 1: Write the failing store and runtime tests**

```go
func TestStoreListEventsReturnsNewestFirst(t *testing.T) {
	t.Parallel()

	store, err := Open("file:test-events?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.UpsertDisk(ctx, domain.Disk{ID: "disk-sda", Name: "sda", Path: "/dev/sda"}); err != nil {
		t.Fatalf("UpsertDisk() error = %v", err)
	}

	if err := store.AppendEvent(ctx, domain.Event{DiskID: "disk-sda", Kind: "older", Message: "older", CreatedAt: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}
	if err := store.AppendEvent(ctx, domain.Event{DiskID: "disk-sda", Kind: "newer", Message: "newer", CreatedAt: time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}

	events, err := store.ListEvents(ctx, 10)
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	if events[0].Kind != "newer" {
		t.Fatalf("events[0].Kind = %q, want %q", events[0].Kind, "newer")
	}
}

func TestRuntimeBuildsTopologyFromDiscoverySnapshot(t *testing.T) {
	t.Parallel()

	runtime := newTestRuntime(t, domain.DiscoverySnapshot{
		Disks: []domain.Disk{
			{ID: "disk-sda", Name: "sda", Path: "/dev/sda", Model: "WD Red", Rotational: true},
		},
		Mounts: []domain.Mount{
			{DiskID: "disk-sda", Source: "/dev/sda1", Target: "/srv/data"},
		},
	})

	graph, err := runtime.Topology(context.Background())
	if err != nil {
		t.Fatalf("Topology() error = %v", err)
	}

	if len(graph.Nodes) != 2 {
		t.Fatalf("len(graph.Nodes) = %d, want 2", len(graph.Nodes))
	}
	if len(graph.Edges) != 1 {
		t.Fatalf("len(graph.Edges) = %d, want 1", len(graph.Edges))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/store/sqlite ./internal/app`

Expected: FAIL with missing `ListEvents`, missing runtime helpers, or missing topology behavior.

- [ ] **Step 3: Add the minimal runtime and persistence code**

```go
type EventStore interface {
	AppendEvent(context.Context, domain.Event) error
	ListEvents(context.Context, int) ([]domain.Event, error)
}

type Runtime struct {
	config   config.Config
	discover discovery.Service
	events   EventStore

	mu          sync.RWMutex
	snapshot    domain.DiscoverySnapshot
	refreshedAt time.Time
	tasks       map[string]DiskTask
}

type DiskTask struct {
	DiskID      string
	Kind        string
	State       string
	RequestedAt time.Time
	ExecuteAt   time.Time
	LastError   string
}

func (r *Runtime) Topology(ctx context.Context) (domain.TopologyGraph, error) {
	snapshot, err := r.Snapshot(ctx)
	if err != nil {
		return domain.TopologyGraph{}, err
	}

	nodes := make([]domain.TopologyNode, 0, len(snapshot.Disks)+len(snapshot.Mounts))
	edges := make([]domain.TopologyEdge, 0, len(snapshot.Mounts))

	for _, disk := range snapshot.Disks {
		nodes = append(nodes, domain.TopologyNode{ID: disk.ID, Kind: "disk", Label: disk.Path})
	}
	for _, mount := range snapshot.Mounts {
		mountID := "mount-" + mount.Target
		nodes = append(nodes, domain.TopologyNode{ID: mountID, Kind: "mount", Label: mount.Target})
		if mount.DiskID != "" {
			edges = append(edges, domain.TopologyEdge{From: mount.DiskID, To: mountID})
		}
	}

	return domain.TopologyGraph{Nodes: nodes, Edges: edges}, nil
}

func (s *Store) ListEvents(ctx context.Context, limit int) ([]domain.Event, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, disk_id, kind, message, created_at FROM events ORDER BY created_at DESC, id DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(&event.ID, &event.DiskID, &event.Kind, &event.Message, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/store/sqlite ./internal/app`

Expected: PASS for the new persistence and runtime tests.

- [ ] **Step 5: Commit**

```bash
git add internal/domain/types.go internal/store/sqlite/store.go internal/store/sqlite/store_test.go internal/app/runtime.go internal/app/runtime_test.go
git commit -m "feat: add runtime snapshot and event helpers"
```

### Task 2: Wire Live API Endpoints And Disk Actions

**Files:**
- Modify: `internal/api/router.go`
- Modify: `internal/api/disks.go`
- Modify: `internal/api/topology.go`
- Modify: `internal/api/settings.go`
- Modify: `internal/api/events.go`
- Modify: `internal/app/app.go`
- Create: `internal/api/live_api_test.go`

- [ ] **Step 1: Write the failing API tests**

```go
func TestDisksEndpointReturnsDiscoveredDisk(t *testing.T) {
	t.Parallel()

	runtime := newStubRuntime()
	runtime.disks = []apiDisk{
		{ID: "disk-sda", Name: "sda", Model: "WD Red", Path: "/dev/sda", Unsupported: false},
	}

	router := api.NewRouter(api.Dependencies{
		TokenPlaintext: "dev-token",
		Runtime:        runtime,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/disks", nil)
	req.AddCookie(&http.Cookie{Name: "diskhm_session", Value: "dev-session"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"disk-sda"`) {
		t.Fatalf("body = %q, want disk payload", rec.Body.String())
	}
}

func TestSleepNowEndpointRunsRuntimeAction(t *testing.T) {
	t.Parallel()

	runtime := newStubRuntime()
	router := api.NewRouter(api.Dependencies{
		TokenPlaintext: "dev-token",
		Runtime:        runtime,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/disks/disk-sda/sleep-now", nil)
	req.AddCookie(&http.Cookie{Name: "diskhm_session", Value: "dev-session"})
	req.Header.Set("X-CSRF-Token", "dev-csrf")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if runtime.sleepNowDiskID != "disk-sda" {
		t.Fatalf("sleepNowDiskID = %q, want %q", runtime.sleepNowDiskID, "disk-sda")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/api ./internal/app`

Expected: FAIL with missing runtime dependency, wrong placeholder JSON, or missing action routes.

- [ ] **Step 3: Implement live handlers with runtime-backed dependencies**

```go
type Runtime interface {
	ListDisks(context.Context) ([]DiskView, error)
	Topology(context.Context) (domain.TopologyGraph, error)
	Settings(context.Context) (SettingsView, error)
	ListEvents(context.Context, int) ([]domain.Event, error)
	SubscribeEvents() (<-chan domain.Event, func())
	SleepNow(context.Context, string) error
	SleepAfter(context.Context, string, int) error
	CancelSleep(context.Context, string) error
	RefreshSafe(context.Context, string) error
	RefreshWake(context.Context, string) error
}

type Dependencies struct {
	TokenPlaintext string
	Runtime        Runtime
}

func registerDiskRoutes(mux *http.ServeMux, deps Dependencies) {
	mux.Handle("GET /api/disks", requireSession(http.HandlerFunc(listDisksHandler(deps))))
	mux.Handle("POST /api/disks/{id}/sleep-now", requireSession(http.HandlerFunc(sleepNowHandler(deps))))
	mux.Handle("POST /api/disks/{id}/sleep-after", requireSession(http.HandlerFunc(sleepAfterHandler(deps))))
	mux.Handle("POST /api/disks/{id}/cancel-sleep", requireSession(http.HandlerFunc(cancelSleepHandler(deps))))
	mux.Handle("POST /api/disks/{id}/refresh-safe", requireSession(http.HandlerFunc(refreshSafeHandler(deps))))
	mux.Handle("POST /api/disks/{id}/refresh-wake", requireSession(http.HandlerFunc(refreshWakeHandler(deps))))
}

func listDisksHandler(deps Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		disks, err := deps.Runtime.ListDisks(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": disks})
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/api ./internal/app`

Expected: PASS for live handler and action route coverage.

- [ ] **Step 5: Commit**

```bash
git add internal/api/router.go internal/api/disks.go internal/api/topology.go internal/api/settings.go internal/api/events.go internal/api/live_api_test.go internal/app/app.go
git commit -m "feat: wire live disk API and actions"
```

### Task 3: Replace Scaffold Frontend Data With Live Queries And Actions

**Files:**
- Modify: `web/src/app/routes.tsx`
- Modify: `web/src/app/App.test.tsx`
- Create: `web/src/features/disks/useDisksQuery.ts`
- Create: `web/src/features/disks/useDiskActions.ts`
- Modify: `web/src/features/disks/DiskTablePage.tsx`
- Modify: `web/src/features/disks/DiskTablePage.test.tsx`
- Modify: `web/src/features/events/EventsPage.tsx`
- Modify: `web/src/features/session/LoginPage.tsx`

- [ ] **Step 1: Write the failing frontend tests**

```tsx
it('renders real disks from the API', async () => {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue(
      new Response('{"items":[{"id":"disk-sda","name":"sda","model":"WD Red","powerState":"unknown","refreshFreshness":"cached","unsupported":false,"mounts":["/srv/data"]}]}', {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    ),
  );

  window.history.pushState({}, '', '/disks');
  renderApp();

  expect(await screen.findByText('WD Red')).toBeInTheDocument();
  expect(screen.queryByText('View scaffold pending backend integration.')).not.toBeInTheDocument();
});

it('posts sleep-now and refreshes the disk query', async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(new Response('{"items":[{"id":"disk-sda","name":"sda","model":"WD Red","powerState":"unknown","refreshFreshness":"cached","unsupported":false,"mounts":[]}]}', { status: 200, headers: { 'Content-Type': 'application/json' } }))
    .mockResolvedValueOnce(new Response(null, { status: 204 }))
    .mockResolvedValueOnce(new Response('{"items":[{"id":"disk-sda","name":"sda","model":"WD Red","powerState":"sleeping","refreshFreshness":"cached","unsupported":false,"mounts":[]}]}', { status: 200, headers: { 'Content-Type': 'application/json' } }));

  vi.stubGlobal('fetch', fetchMock);
  window.history.pushState({}, '', '/disks');
  renderApp();

  fireEvent.click(await screen.findByRole('button', { name: 'Sleep now' }));

  await waitFor(() => {
    expect(fetchMock).toHaveBeenCalledWith('/api/disks/disk-sda/sleep-now', expect.objectContaining({ method: 'POST' }));
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `npx vitest run src/app/App.test.tsx src/features/disks/DiskTablePage.test.tsx`

Expected: FAIL because `/disks` is not live yet and action hooks do not exist.

- [ ] **Step 3: Implement live queries, actions, and route wiring**

```tsx
export function useDisksQuery() {
  return useQuery({
    queryKey: ['disks'],
    queryFn: async () => {
      const payload = await http.json<{ items: DiskListItem[] }>('/api/disks');
      return payload.items;
    },
  });
}

export function useDiskActions() {
  const queryClient = useQueryClient();

  async function postAction(path: string) {
    const response = await http.request(path, { method: 'POST' });
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    await queryClient.invalidateQueries({ queryKey: ['disks'] });
    await queryClient.invalidateQueries({ queryKey: ['topology'] });
    await queryClient.invalidateQueries({ queryKey: ['events'] });
  }

  return {
    sleepNow: (diskID: string) => postAction(`/api/disks/${diskID}/sleep-now`),
    refreshSafe: (diskID: string) => postAction(`/api/disks/${diskID}/refresh-safe`),
    refreshWake: (diskID: string) => postAction(`/api/disks/${diskID}/refresh-wake`),
  };
}
```

- [ ] **Step 4: Run tests and production build**

Run: `npx vitest run`

Expected: PASS

Run: `npm run build`

Expected: PASS and updated assets in `internal/webassets/dist/`

- [ ] **Step 5: Commit**

```bash
git add web/src/app/routes.tsx web/src/app/App.test.tsx web/src/features/disks/useDisksQuery.ts web/src/features/disks/useDiskActions.ts web/src/features/disks/DiskTablePage.tsx web/src/features/disks/DiskTablePage.test.tsx web/src/features/events/EventsPage.tsx internal/webassets/dist
git commit -m "feat: connect live frontend disk pages"
```

### Task 4: Add End-To-End Runtime Verification And Host Checks

**Files:**
- Modify: `README.md`
- Modify: `docs/manual-test/diskhm-mvp-smoke.md`

- [ ] **Step 1: Add the failing verification expectations to docs**

```md
- [ ] Log in with the configured token and confirm the app redirects to `/topology`.
- [ ] Open `/disks` and confirm at least one discovered disk row appears on a populated host.
- [ ] Trigger `Refresh (wake disk)` only on an explicitly supported disk and confirm a new event appears.
- [ ] Trigger `Sleep now` on a supported HDD and confirm the action result appears in the events list.
```

- [ ] **Step 2: Run the existing automated verification before doc updates**

Run: `npx vitest run`

Expected: PASS

Run: `npm run build`

Expected: PASS

- [ ] **Step 3: Update verification docs with the real runtime flow**

```md
## Runtime

- [ ] Start `diskhmd daemon --config /etc/diskhm/config.yaml`.
- [ ] Open `http://127.0.0.1:9789/` and log in with the configured token.
- [ ] Confirm `/topology` shows live route content rather than the scaffold message.
- [ ] Confirm `/events` can show persisted or streamed action events.
- [ ] Confirm `/settings` shows the configured quiet grace seconds.
```

- [ ] **Step 4: Re-run the final automated checks**

Run: `npx vitest run`

Expected: PASS

Run: `npm run build`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add README.md docs/manual-test/diskhm-mvp-smoke.md
git commit -m "docs: update live runtime verification steps"
```
