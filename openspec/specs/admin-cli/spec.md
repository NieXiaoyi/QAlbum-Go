# Admin CLI

## Purpose

TBD

## Requirements

### Requirement: Admin CLI user management
The system SHALL provide a command-line admin tool for manually managing users in the `users` table.

#### Scenario: Add user by OpenID
- **WHEN** an administrator runs `./qalbum-admin user add --openid <openid>`
- **THEN** the system inserts the OpenID into the `users` table
- **AND** outputs the newly created user ID

#### Scenario: List all users
- **WHEN** an administrator runs `./qalbum-admin user list`
- **THEN** the system outputs all users with their `id`, `openid`, and `created_at`

### Requirement: New member onboarding workflow
The system SHALL support a manual onboarding workflow where new members are added by the administrator after a failed login attempt.

#### Scenario: New member login rejected and logged
- **WHEN** an unregistered user attempts to log in via `/auth/login`
- **THEN** the system logs the rejected OpenID with an `[AUTH]` prefix
- **AND** returns HTTP 403

#### Scenario: Admin adds member and re-login succeeds
- **WHEN** the administrator extracts the OpenID from logs and runs `./qalbum-admin user add --openid <openid>`
- **AND** the user attempts to log in again
- **THEN** the system successfully authenticates the user and returns a JWT token
