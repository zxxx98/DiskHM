# DiskHM Live Data And Actions Design

Date: 2026-05-01

## Goal

Replace the remaining placeholder backend and frontend behavior with working live data and disk actions for the current DiskHM MVP:

- return real disk, topology, settings, and event data from the daemon
- show that real data in the web UI instead of scaffold views
- wire disk actions to the existing scheduler, refresh, and power-control services

This spec extends the existing MVP design in `2026-04-29-diskhm-design.md` and focuses on implementation shape for the current codebase.

## Scope

This change includes:

- live `GET /api/disks`
- live `GET /api/topology`
- live `GET /api/settings`
- live `GET /api/events`
- live `GET /api/events/stream`
- working `POST /api/disks/{id}/sleep-now`
- working `POST /api/disks/{id}/sleep-after`
- working `POST /api/disks/{id}/cancel-sleep`
- working `POST /api/disks/{id}/refresh-safe`
- working `POST /api/disks/{id}/refresh-wake`
- web routes that render real data instead of shell placeholders

This change does not include:

- persistent job recovery after daemon restart
- full mdadm/LVM/ZFS topology modeling
- SMART collection
- writeable settings UI beyond current read-only behavior
- multi-user auth or expanded session model

## Current State

The codebase already contains useful building blocks:

- `internal/discovery` can read disks and mounts from sysfs, udev, and mountinfo
- `internal/scheduler` can perform quiet-window gated sleep-now execution
- `internal/refresh` can model safe and wake refresh behavior
- `internal/power` contains `hdparm`-based standby and state helpers
- frontend pages for topology, events, settings, and disks already exist

The remaining gap is wiring:

- API handlers still return hard-coded placeholder JSON
- frontend routes still mostly reflect scaffold behavior
- no runtime application service ties discovery, actions, and event recording together

## Recommended Approach

Implement a small runtime service layer inside the daemon and make the API handlers depend on it.

Why this approach:

- it reuses the existing discovery, scheduler, refresh, and power packages
- it keeps HTTP handlers thin and testable
- it avoids hard-coding operational logic into route handlers
- it gives the frontend stable API responses without redesigning the domain model

## Runtime Design

Add an application runtime object owned by `internal/app` that wires together:

- config
- discovery service
- disk action services
- in-memory task state
- SQLite-backed event persistence
- SSE subscribers

The runtime is responsible for:

- refreshing and caching discovery snapshots
- mapping discovered disks into API response models
- exposing topology and settings data
- running disk actions and recording events
- broadcasting new events to SSE listeners

## Data Flow

### Discovery

Discovery remains sysfs/udev/mountinfo based.

Request flow:

1. API asks runtime for current snapshot.
2. Runtime serves a recent cached snapshot when safe to do so.
3. Runtime refreshes discovery on demand for safe refresh requests.
4. Wake refresh may additionally call wake-capable probing paths for a specific disk.

Initial caching rule:

- keep one in-memory snapshot with a refresh timestamp
- treat normal page loads as safe reads from the cached snapshot
- refresh snapshot on explicit `refresh-safe`
- refresh snapshot after action completion when the result may change device state

### Disk Actions

Action flow:

1. Client sends disk action request.
2. Handler resolves disk ID through the current discovery snapshot.
3. Runtime validates device support and current task state.
4. Runtime runs the action or schedules it in memory.
5. Runtime records an event and updates in-memory task state.
6. Runtime broadcasts the event through SSE.

## API Design

### `GET /api/disks`

Return a real disk list derived from discovery plus in-memory task state.

Response fields should cover the current frontend table:

- `id`
- `name`
- `model`
- `path`
- `powerState`
- `refreshFreshness`
- `unsupported`
- `mounts`
- `task`

Simplifying rules for MVP:

- `powerState` may initially be `unknown` unless a safe state source exists
- rotational disks are candidates for HDD sleep actions
- non-rotational disks render as unsupported for HDD sleep
- USB/SATA bridges without a safe command path remain unsupported

### `GET /api/topology`

Build a read-only graph from the discovery snapshot:

- one node per disk
- one node per mount target
- edge from disk to mount when `mount.DiskID` resolves

This is intentionally simpler than the long-term design, but much better than a placeholder page and consistent with current available data.

### `GET /api/settings`

Return live config-backed settings for the current read-only settings page:

- `quiet_grace_seconds`
- optionally `listen_addr` later if the UI starts showing it

### `GET /api/events`

Return recent persisted events from SQLite in reverse chronological order.

### `GET /api/events/stream`

Keep the existing SSE endpoint but attach it to a real broadcaster so the events page can reflect new actions.

### Disk Action Endpoints

Implement:

- `POST /api/disks/{id}/sleep-now`
- `POST /api/disks/{id}/sleep-after`
- `POST /api/disks/{id}/cancel-sleep`
- `POST /api/disks/{id}/refresh-safe`
- `POST /api/disks/{id}/refresh-wake`

Behavior:

- `sleep-now` uses `internal/scheduler`
- `sleep-after` stores an in-memory delayed task and later enters the same execution path as `sleep-now`
- `cancel-sleep` removes that in-memory task if present
- `refresh-safe` refreshes snapshot only through safe discovery paths
- `refresh-wake` explicitly allows wake-capable probe behavior for the selected disk

## Task Model

Use an in-memory per-disk task map in the first implementation.

Tracked fields:

- disk ID
- task kind
- state
- requested time
- scheduled wake time for delayed sleep
- last error message

States:

- `scheduled`
- `waiting_idle`
- `executing`
- `sleeping`
- `failed`
- `canceled`

This matches the existing product vocabulary closely enough for UI rendering and event generation.

Persistence note:

- tasks do not survive daemon restart in this phase
- persisted recovery can be added later without invalidating the API shape

## Event Model

Events should be recorded for:

- daemon-initiated snapshot refresh
- sleep-now requested
- sleep-after requested
- sleep canceled
- sleep execution started
- sleep execution succeeded
- sleep execution failed
- wake refresh requested
- safe refresh requested

Each event should include:

- `disk_id` when applicable
- `kind`
- human-readable `message`
- `created_at`

## Frontend Design

### App Routing

Keep the new router-based app structure and replace remaining scaffold route behavior with live route components.

### Topology Page

Use live `GET /api/topology`.

If empty:

- render the current empty state text
- do not fall back to scaffold wording

### Events Page

Use:

- initial load from `GET /api/events`
- live append from SSE stream

If empty:

- render a real empty state like `No events received yet.`

### Settings Page

Use live `GET /api/settings`.

Remain read-only in this phase.

### Disk Inventory Page

Wire the existing disk table components to live `GET /api/disks`.

Actions should call the new endpoints and then:

- invalidate disks query
- invalidate topology query when appropriate
- append or receive live events

## Error Handling

Use structured JSON errors for action failures.

Initial codes:

- `disk_not_found`
- `unsupported_device`
- `disk_busy`
- `task_conflict`
- `refresh_requires_wake`
- `command_failed`

Frontend behavior:

- show per-action failure near the disk interaction
- keep the rest of the page usable
- let follow-up refresh recover state naturally

## Testing Strategy

### Backend

Add tests for:

- runtime snapshot mapping from discovery data to API payloads
- topology graph generation from disks and mounts
- settings response from config
- event listing and SSE broadcast
- sleep-now routing into scheduler
- sleep-after scheduling and cancel behavior
- unsupported-device rejection

### Frontend

Add tests for:

- `/topology` renders real route data, not scaffold text
- events page loads initial events and handles empty state
- settings page loads live quiet grace seconds
- disk list renders API-provided disks and action buttons
- action success invalidates data and action failure surfaces an error

### Manual Verification

On the Linux host:

- log in successfully
- open topology and confirm no scaffold message appears
- confirm disks page shows real disks
- issue safe refresh and verify page updates without obvious wake side-effects
- issue sleep-now for a supported HDD and observe event creation

## Risks And Mitigations

Risk: power-state detection may still be incomplete.
Mitigation: surface `unknown` honestly instead of faking active/sleeping.

Risk: delayed sleep tasks disappear on restart.
Mitigation: clearly keep this behavior in-memory for this phase and log cancellation on startup later if needed.

Risk: USB bridge compatibility is inconsistent.
Mitigation: keep unsupported classification conservative and require explicit wake actions.

Risk: live routes increase coupling between frontend and API payloads.
Mitigation: add focused API-shape tests and keep route loaders thin.

## Acceptance Criteria

- Login succeeds and navigates into the app.
- Topology route renders a real page backed by `GET /api/topology`.
- Scaffold placeholder text is gone from live routes.
- Disk list comes from real discovery data.
- Settings page reads live config-backed values.
- Events page shows persisted events and receives live ones.
- Sleep and refresh actions hit real backend logic and emit events.
- Frontend tests pass.
- Web production build passes.
