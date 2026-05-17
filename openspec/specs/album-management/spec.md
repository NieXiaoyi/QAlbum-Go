# Album Management

## Purpose

TBD

## Requirements

### Requirement: Create album
The system SHALL allow any space member to create an album within a family space.

#### Scenario: Successful creation
- **WHEN** a space member POSTs to `/spaces/{space_id}/albums` with a `name`
- **THEN** the system creates an album record associated with the space
- **AND** returns the created album with HTTP 201

### Requirement: List albums
The system SHALL allow any space member to list albums in a space, excluding soft-deleted ones.

#### Scenario: List albums
- **WHEN** a space member GETs `/spaces/{space_id}/albums`
- **THEN** the system returns a paginated list of albums where `deleted_at IS NULL`

### Requirement: Get album details
The system SHALL allow any space member to view a specific album's details.

#### Scenario: Get album
- **WHEN** a space member GETs `/spaces/{space_id}/albums/{album_id}`
- **THEN** the system returns the album record

### Requirement: Rename album
The system SHALL allow any space member to rename an album.

#### Scenario: Successful rename
- **WHEN** a space member PUTs to `/spaces/{space_id}/albums/{album_id}` with a new `name`
- **THEN** the system updates the album's `name` and `updated_at`
- **AND** returns the updated album

### Requirement: Delete album (soft delete)
The system SHALL allow any space member to soft-delete an album. Physical files are NOT affected.

#### Scenario: Soft delete
- **WHEN** a space member DELETEs `/spaces/{space_id}/albums/{album_id}`
- **THEN** the system sets `albums.deleted_at` to the current timestamp
- **AND** files within the album become invisible to users (filtered via JOIN)
- **AND** returns HTTP 204

#### Scenario: Physical cleanup after 30 days
- **WHEN** the daily cleanup cron job runs
- **AND** finds albums with `deleted_at < datetime('now', '-30 days')`
- **THEN** the system physically deletes all files under those albums (raw, thumb, cover, backup)
- **AND** hard-deletes the album record from the database
