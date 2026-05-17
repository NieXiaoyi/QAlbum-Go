# Space Management

## Purpose

TBD

## Requirements

### Requirement: Create family space
The system SHALL allow an authenticated user to create a family space and automatically become its admin.

#### Scenario: Successful creation
- **WHEN** an authenticated user POSTs to `/spaces` with `name` and `quota_bytes`
- **THEN** the system creates a space record, sets `owner_id` to the user's ID
- **AND** inserts the user into `space_members` with role `admin`
- **AND** returns the created space with HTTP 201

### Requirement: List and view family spaces
The system SHALL allow a user to list all spaces they are a member of and view individual space details.

#### Scenario: List spaces
- **WHEN** an authenticated user GETs `/spaces`
- **THEN** the system returns a paginated list of all spaces where the user is a member

#### Scenario: View space details
- **WHEN** an authenticated member GETs `/spaces/{space_id}`
- **THEN** the system returns the space details including `member_count` and `pending_backup_count`

### Requirement: Update family space settings
The system SHALL allow the space admin to update space name, quota, and backup path.

#### Scenario: Admin updates settings
- **WHEN** the space admin PUTs to `/spaces/{space_id}` with updated fields
- **THEN** the system updates the space record
- **AND** if `backup_path` is changed to a non-empty value, marks all files in the space as `pending` backup status

#### Scenario: Non-admin update rejected
- **WHEN** a non-admin member PUTs to `/spaces/{space_id}`
- **THEN** the system returns HTTP 403 Forbidden

### Requirement: Delete family space
The system SHALL allow the space admin to permanently delete a space and all its associated data.

#### Scenario: Admin deletes space
- **WHEN** the space admin DELETEs `/spaces/{space_id}`
- **THEN** the system physically deletes all files under `/data/photos/<space_id>/`
- **AND** cascades deletion of albums, files, members, and invite tokens from the database
- **AND** returns HTTP 204

### Requirement: Generate invite token
The system SHALL allow the space admin to generate a one-time invite token for other users to join the space.

#### Scenario: Admin generates invite
- **WHEN** the space admin POSTs to `/spaces/{space_id}/invite` with optional `expire_hours`
- **THEN** the system creates an `invite_tokens` record with a random token string
- **AND** returns the `invite_code` and `expires_at`

### Requirement: Join space by invite token
The system SHALL allow an authenticated user to join a space using a valid invite token.

#### Scenario: Valid token join
- **WHEN** an authenticated user POSTs to `/spaces/join` with an `invite_code`
- **AND** the token exists, is unused, and not expired
- **THEN** the system marks the token as used
- **AND** inserts the user into `space_members` with role `member`
- **AND** returns the space details

### Requirement: Manage space members
The system SHALL allow the space admin to list members and remove non-admin members.

#### Scenario: List members
- **WHEN** an authenticated member GETs `/spaces/{space_id}/members`
- **THEN** the system returns the list of members with `user_id`, `role`, and `joined_at`

#### Scenario: Remove member
- **WHEN** the space admin DELETEs `/spaces/{space_id}/members/{user_id}`
- **AND** the target user is not the admin themselves
- **THEN** the system removes the user from `space_members`
- **AND** returns HTTP 204
