CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    openid TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE spaces (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    owner_id INTEGER NOT NULL,
    quota_bytes INTEGER NOT NULL,
    backup_path TEXT,
    backup_status TEXT DEFAULT 'unavailable' CHECK (backup_status IN ('ok', 'pending', 'unavailable')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE space_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    space_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('admin', 'member')),
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(space_id, user_id),
    FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE albums (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    space_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    deleted_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE
);

CREATE TABLE files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    space_id INTEGER NOT NULL,
    album_id INTEGER NOT NULL,
    filename TEXT NOT NULL,
    file_type TEXT NOT NULL CHECK (file_type IN ('image', 'video')),
    file_size INTEGER NOT NULL,
    mime_type TEXT NOT NULL,
    storage_path TEXT NOT NULL,
    thumbnail_path TEXT,
    cover_path TEXT,
    phash TEXT,
    width INTEGER,
    height INTEGER,
    duration INTEGER,
    uploaded_by INTEGER NOT NULL,
    uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    backup_status TEXT DEFAULT 'pending' CHECK (backup_status IN ('ok', 'pending', 'unavailable')),
    FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE,
    FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE SET NULL,
    FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE invite_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    space_id INTEGER NOT NULL,
    token TEXT UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    used_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (space_id) REFERENCES spaces(id) ON DELETE CASCADE
);

CREATE INDEX idx_space_members_space_id ON space_members(space_id);
CREATE INDEX idx_space_members_user_id ON space_members(user_id);
CREATE INDEX idx_albums_space_id ON albums(space_id);
CREATE INDEX idx_files_space_id ON files(space_id);
CREATE INDEX idx_files_album_id ON files(album_id);
CREATE INDEX idx_files_uploaded_by ON files(uploaded_by);
CREATE INDEX idx_files_phash ON files(phash);
CREATE INDEX idx_files_deleted_at ON files(deleted_at);
CREATE INDEX idx_invite_tokens_token ON invite_tokens(token);
CREATE INDEX idx_invite_tokens_expires_at ON invite_tokens(expires_at);
