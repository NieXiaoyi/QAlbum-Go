# File Management

## Purpose

TBD

## Requirements

### Requirement: Upload file
The system SHALL allow any space member to upload an image or video file to an album, subject to space quota.

#### Scenario: Successful upload
- **WHEN** a space member POSTs a multipart file to `/spaces/{space_id}/albums/{album_id}/files`
- **THEN** the system checks the space quota using a `BEGIN IMMEDIATE` transaction
- **AND** computes the file MD5, extracts width/height (images) or duration (videos)
- **AND** saves the file to `/data/photos/<space_id>/raw/<YYYY/MM>/<timestamp>_<rand>.<ext>`
- **AND** creates a `files` database record
- **AND** enqueues asynchronous tasks for thumbnail/cover generation and backup
- **AND** returns the file record with HTTP 201

#### Scenario: Quota exceeded
- **WHEN** a space member attempts to upload a file that would exceed `quota_bytes`
- **THEN** the system returns HTTP 413 with a quota exceeded error

#### Scenario: File too large
- **WHEN** a uploaded file exceeds `max_upload_size` (200MB)
- **THEN** the system returns HTTP 413 with a file too large error

### Requirement: List files in album
The system SHALL allow any space member to list files in an album, paginated and filterable by type.

#### Scenario: List files
- **WHEN** a space member GETs `/spaces/{space_id}/albums/{album_id}/files?type=all&page=1&page_size=20`
- **THEN** the system returns a paginated list of files where `deleted_at IS NULL`
- **AND** excludes files from soft-deleted albums

### Requirement: Get file details
The system SHALL allow any space member to view a specific file's metadata.

#### Scenario: Get file
- **WHEN** a space member GETs `/spaces/{space_id}/files/{file_id}`
- **THEN** the system returns the file record

### Requirement: Rename or move file
The system SHALL allow any space member to rename a file or move it to another album within the same space.

#### Scenario: Rename file
- **WHEN** a space member PUTs to `/spaces/{space_id}/files/{file_id}` with `filename`
- **THEN** the system updates the file's `filename`

#### Scenario: Move file to another album
- **WHEN** a space member PUTs to `/spaces/{space_id}/files/{file_id}` with `album_id`
- **AND** the target album exists in the same space
- **THEN** the system updates `files.album_id` only (no physical file movement)

### Requirement: Delete file (soft delete to recycle bin)
The system SHALL allow any space member to soft-delete a file, moving it to the recycle bin.

#### Scenario: Soft delete
- **WHEN** a space member DELETEs `/spaces/{space_id}/files/{file_id}`
- **THEN** the system sets `files.deleted_at` to the current timestamp
- **AND** the file becomes invisible in normal album views
- **AND** returns HTTP 204

### Requirement: Download original file
The system SHALL allow any space member to download the original file.

#### Scenario: Download
- **WHEN** a space member GETs `/spaces/{space_id}/files/{file_id}/download`
- **THEN** the system streams the original file with `Content-Type: application/octet-stream`

### Requirement: Get thumbnail or cover
The system SHALL allow any space member to retrieve a file's thumbnail (images) or cover (videos).

#### Scenario: Thumbnail available
- **WHEN** a space member GETs `/spaces/{space_id}/files/{file_id}/thumbnail`
- **AND** the thumbnail/cover exists
- **THEN** the system returns the image with `Content-Type: image/jpeg`

#### Scenario: Thumbnail pending
- **WHEN** a space member GETs `/spaces/{space_id}/files/{file_id}/thumbnail`
- **AND** the thumbnail/cover is not yet generated
- **THEN** the system returns HTTP 202 with a generation-in-progress message
