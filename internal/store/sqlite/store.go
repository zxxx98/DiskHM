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
		INSERT INTO disks (id, name, model, path)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			model = excluded.model,
			path = excluded.path
		`,
		disk.ID,
		disk.Name,
		disk.Model,
		disk.Path,
	)
	return err
}

func (s *Store) ListDisks(ctx context.Context) ([]domain.Disk, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, name, model, path FROM disks ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var disks []domain.Disk
	for rows.Next() {
		var disk domain.Disk
		if err := rows.Scan(&disk.ID, &disk.Name, &disk.Model, &disk.Path); err != nil {
			return nil, err
		}
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
		`INSERT INTO events (disk_id, kind) VALUES (?, ?)`,
		event.DiskID,
		event.Kind,
	)
	return err
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
