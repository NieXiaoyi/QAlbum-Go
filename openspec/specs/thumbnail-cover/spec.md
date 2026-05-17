# Thumbnail & Cover

## Purpose

TBD

## Requirements

### Requirement: Generate image thumbnail asynchronously
The system SHALL generate a JPEG thumbnail for every uploaded image asynchronously after upload completes.

#### Scenario: Thumbnail generation
- **WHEN** an image file is successfully uploaded and a `ThumbnailTask` is enqueued
- **THEN** a worker generates a 400px max-dimension JPEG thumbnail using the `imaging` library with Lanczos resampling and 85% quality
- **AND** saves it to `/data/photos/<space_id>/thumb/<YYYY/MM>/<basename>_thumb.jpg`
- **AND** updates `files.thumb_path` in the database

### Requirement: Extract video cover asynchronously
The system SHALL extract the first frame of every uploaded video as a JPEG cover asynchronously.

#### Scenario: Cover extraction
- **WHEN** a video file is successfully uploaded and a `CoverTask` is enqueued
- **THEN** a worker calls FFmpeg with `-ss 00:00:01 -vframes 1 -q:v 2`
- **AND** limits concurrent FFmpeg processes to a configurable maximum (default 2)
- **AND** saves the cover to `/data/photos/<space_id>/cover/<YYYY/MM>/<basename>_cover.jpg`
- **AND** updates `files.cover_path` in the database

#### Scenario: FFmpeg timeout
- **WHEN** an FFmpeg process exceeds 30 seconds
- **THEN** the system kills the process
- **AND** logs the failure without blocking other tasks

### Requirement: Audit missing metadata
The system SHALL periodically audit files missing thumbnail, cover, or phash metadata and re-enqueue tasks.

#### Scenario: Audit discovers missing thumbnail
- **WHEN** the audit cron job runs and finds an image with `thumb_path IS NULL` and `uploaded_at` older than `upload_grace_hours`
- **THEN** the system enqueues a new `ThumbnailTask` for that file

#### Scenario: Audit discovers missing cover
- **WHEN** the audit cron job finds a video with `cover_path IS NULL` and `uploaded_at` older than `upload_grace_hours`
- **THEN** the system enqueues a new `CoverTask` for that file
