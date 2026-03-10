# API Documentation - Auth & Monitoring

This API supports hybrid authentication:

1. **JWT (Users):** Use `Authorization: Bearer <token>`
2. **API Key (Services):** Use `X-API-Key: <key>`

## Endpoints

### 1. Authentication

#### POST `/auth/login`

Authenticates a user and returns a token pair.

- **Request Body:**

  ```json
  {
    "uid": "user123",
    "secret": "password"
  }
  ```

- **Response:**

  ```json
  {
    "access_token": "...",
    "refresh_token": "...",
    "expires_at": 123456789
  }
  ```

#### POST `/auth/refresh`

Refreshes an expired access token using a refresh token.

- **Request Body:**

  ```json
  {
    "refresh_token": "..."
  }
  ```

### 2. Monitoring

#### GET `/api/v1/pools`

Returns the status of monitored liquidity pools.

- **Auth Required:** JWT or API Key
- **Query Params:**
  - `network` (optional, default: "ethereum")
- **Response:**

  ```json
  {
    "status": "pool monitoring is active for ethereum"
  }
  ```

## Security Headers

- `Authorization: Bearer <JWT>`
- `X-API-Key: <Key>`
