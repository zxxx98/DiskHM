package sqlitestore

import (
	"context"
	_ "embed"
	"errors"
	"math"
	"time"

	"github.com/example/diskhm/internal/domain"
)

//go:embed migrations/001_init.sql
var migrationSQL string

func (s *Store) migrate() error {
	_, err := s.db.Exec(migrationSQL)
	return err
}

func (s *Store) UpsertDisk(ctx context.Context, disk domain.Disk) error {
	if disk.SizeBytes > math.MaxInt64 {
		return errors.New("disk size_bytes exceeds SQLite INTEGER range")
	}

	_, err := s.db.ExecContext(
		ctx,
		`
		INSERT INTO disks (id, name, path, model, serial, transport, size_bytes, rotational)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			path = excluded.path,
			model = excluded.model,
			serial = excluded.serial,
			transport = excluded.transport,
			size_bytes = excluded.size_bytes,
			rotational = excluded.rotational
		`,
		disk.ID,
		disk.Name,
		disk.Path,
		disk.Model,
		disk.Serial,
		disk.Transport,
		int64(disk.SizeBytes),
		boolToInt(disk.Rotational),
	)
	return err
}

func (s *Store) ListDisks(ctx context.Context) ([]domain.Disk, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, name, path, model, serial, transport, size_bytes, rotational FROM disks ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var disks []domain.Disk
	for rows.Next() {
		var disk domain.Disk
		var sizeBytes int64
		var rotational int
		if err := rows.Scan(
			&disk.ID,
			&disk.Name,
			&disk.Path,
			&disk.Model,
			&disk.Serial,
			&disk.Transport,
			&sizeBytes,
			&rotational,
		); err != nil {
			return nil, err
		}
		if sizeBytes < 0 {
			return nil, errors.New("disk size_bytes read from SQLite was negative")
		}
		disk.SizeBytes = uint64(sizeBytes)
		disk.Rotational = rotational != 0
		disks = append(disks, disk)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return disks, nil
}

func (s *Store) AppendEvent(ctx context.Context, event domain.Event) error {
	if event.CreatedAt.IsZero() {
		_, err := s.db.ExecContext(
			ctx,
			`INSERT INTO events (disk_id, kind, message) VALUES (?, ?, ?)`,
			event.DiskID,
			event.Kind,
			event.Message,
		)
		return err
	}

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO events (disk_id, kind, message, created_at) VALUES (?, ?, ?, ?)`,
		event.DiskID,
		event.Kind,
		event.Message,
		event.CreatedAt,
	)
	return err
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
		var createdAt string
		if err := rows.Scan(&event.ID, &event.DiskID, &event.Kind, &event.Message, &createdAt); err != nil {
			return nil, err
		}
		event.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			event.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
			if err != nil {
				event.CreatedAt, err = time.Parse("2006-01-02 15:04:05 -0700 MST", createdAt)
				if err != nil {
					return nil, err
				}
			}
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
