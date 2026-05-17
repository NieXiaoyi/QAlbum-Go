# WeChat Authentication

## Purpose

TBD

## Requirements

### Requirement: User login via WeChat Mini Program
The system SHALL authenticate users using the WeChat Mini Program login flow and issue a JWT token.

#### Scenario: Successful login
- **WHEN** a user sends a valid `code` obtained from `wx.login()` to `/auth/login`
- **THEN** the system calls WeChat `code2session` API to obtain `openid`
- **AND** the system queries the `users` table by `openid`
- **AND** if the user exists, the system generates a JWT token with 6-hour expiration and returns it

#### Scenario: Rejected login for unregistered user
- **WHEN** a user sends a valid `code` to `/auth/login`
- **AND** the corresponding `openid` does NOT exist in the `users` table
- **THEN** the system logs an audit message containing the rejected `openid`
- **AND** returns HTTP 403 with an access denied error

### Requirement: JWT token validation
The system SHALL validate the JWT token on every protected request.

#### Scenario: Valid token
- **WHEN** a request includes a valid Bearer JWT token in the Authorization header
- **THEN** the system extracts `user_id` from the token and allows the request to proceed

#### Scenario: Expired or invalid token
- **WHEN** a request includes an expired or invalid Bearer token
- **THEN** the system returns HTTP 401 Unauthorized
