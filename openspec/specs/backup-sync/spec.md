# Backup & Sync

## Purpose

TBD

## Requirements

### Requirement: Backup file to external storage asynchronously
The system SHALL asynchronously copy uploaded files to an external backup path configured at the space level.

#### Scenario: Successful backup on upload
- **WHEN** a file is uploaded to a space with a non-empty `backup_path`
- **AND** the backup path exists and is writable
- **THEN** a `BackupTask` is enqueued and a worker copies the file to the backup path
- **AND** verifies MD5 equality between source and destination
- **AND** updates `files.backup_status` to `ok`

#### Scenario: Backup path unavailable
- **WHEN** a file is uploaded to a space with a non-empty `backup_path`
- **AND** the backup path does not exist or is not writable
- **THEN** the system sets `files.backup_status` to `pending`
- **AND** increments `spaces.pending_backup_count`
- **AND** does NOT block the upload response

### Requirement: Manual backup sync
The system SHALL allow the space admin to manually trigger a sync of all pending backup files.

#### Scenario: Trigger sync
- **WHEN** the space admin POSTs to `/spaces/{space_id}/backup/sync`
- **THEN** the system queries all files with `backup_status = 'pending'` in the space
- **AND** enqueues `BackupTask` for each
- **AND** returns the count of pending files being synced

### Requirement: Backup path change handling
The system SHALL handle backup path changes by marking all existing files as pending for re-sync.

#### Scenario: Change backup path
- **WHEN** the space admin updates the space `backup_path` to a new non-empty value
- **THEN** the system sets `backup_status = 'pending'` for all files in the space
- **AND** the new uploads and existing files will sync to the new path
