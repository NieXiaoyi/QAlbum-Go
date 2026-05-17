# Duplicate Detection

## Purpose

TBD

## Requirements

### Requirement: Detect visually similar duplicate images
The system SHALL detect duplicate images within a family space using perceptual hashing (pHash).

#### Scenario: Trigger duplicate detection
- **WHEN** a space member GETs `/spaces/{space_id}/duplicates?threshold=10`
- **THEN** the system scans all visible images in the space (cross-album, excluding soft-deleted)
- **AND** computes or retrieves cached 64-bit pHash for each image
- **AND** compares all pairs using Hamming Distance
- **AND** clusters similar images (distance <= threshold) using Union-Find
- **AND** returns groups containing 2 or more files with the maximum distance in each group

### Requirement: Cache pHash values
The system SHALL cache computed pHash values in the database to avoid recomputation.

#### Scenario: Compute and cache pHash
- **WHEN** duplicate detection encounters an image with `phash IS NULL`
- **THEN** the system computes the pHash using DCT-based algorithm
- **AND** stores the hex string in `files.phash`

#### Scenario: Use cached pHash
- **WHEN** duplicate detection encounters an image with `phash IS NOT NULL`
- **THEN** the system uses the cached value directly

### Requirement: Hamming distance calculation
The system SHALL correctly compute Hamming Distance between two 64-bit pHash strings.

#### Scenario: Distance calculation
- **WHEN** the system receives two hex pHash strings
- **THEN** it parses them as uint64 and returns `bits.OnesCount64(h1 ^ h2)`
