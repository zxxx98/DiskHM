package sqlitestore

import (
	"context"
	_ "embed"

	"github.com/example/diskhm/internal/domain"
)

//go:embed migrations/001_init.sql
var migrationSQL string

func (s *Store) migrate() error {
	_, err := s.db.Exec(migrationSQL)
	return err
}

func (s *Store) UpsertDisk(ctx context.Context, disk domain.Disk) error {
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

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
