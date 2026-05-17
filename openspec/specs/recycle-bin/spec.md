# Recycle Bin

## Purpose

TBD

## Requirements

### Requirement: Soft-deleted files enter recycle bin
The system SHALL place soft-deleted files into a space-level recycle bin visible only to the admin.

#### Scenario: View recycle bin
- **WHEN** the space admin GETs `/spaces/{space_id}/recycle`
- **THEN** the system returns a paginated list of files where `deleted_at IS NOT NULL`

#### Scenario: Non-admin view rejected
- **WHEN** a non-admin member GETs `/spaces/{space_id}/recycle`
- **THEN** the system returns HTTP 403 Forbidden

### Requirement: Restore file from recycle bin
The system SHALL allow the space admin to restore a file from the recycle bin.

#### Scenario: Restore file
- **WHEN** the space admin POSTs to `/spaces/{space_id}/recycle/{file_id}/restore`
- **THEN** the system sets `files.deleted_at` to NULL
- **AND** if the original album is soft-deleted, moves the file to the first available album or creates an "未分类" album
- **AND** returns the restored file record

### Requirement: Clear recycle bin
The system SHALL allow the space admin to permanently delete all files in the recycle bin.

#### Scenario: Clear recycle bin
- **WHEN** the space admin POSTs to `/spaces/{space_id}/recycle/clear`
- **THEN** the system physically deletes all raw files, thumbnails, covers, and backup copies for files in the recycle bin
- **AND** hard-deletes the file records from the database
- **AND** returns HTTP 204

### Requirement: Automatic cleanup of expired recycled files
The system SHALL automatically physically delete files that have been in the recycle bin for more than 30 days.

#### Scenario: Daily cleanup job
- **WHEN** the daily cleanup cron job runs
- **AND** finds files with `deleted_at IS NOT NULL AND deleted_at < datetime('now', '-30 days')`
- **THEN** the system physically deletes their raw, thumb, cover, and backup files
- **AND** hard-deletes the records from the database
