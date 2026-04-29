CREATE TABLE IF NOT EXISTS disks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    model TEXT NOT NULL,
    serial TEXT NOT NULL,
    transport TEXT NOT NULL,
    size_bytes INTEGER NOT NULL,
    rotational INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    disk_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (disk_id) REFERENCES disks(id)
);
