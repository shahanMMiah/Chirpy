# API Endpoints Documentation

This document outlines the available API endpoints and their functionalities.

---

## 1. GET /
**Description:** Health check endpoint.
**Response:** `OK` (plain text)
**Status Codes:**
*   `200 OK`

---

## 2. GET /metrics
**Description:** Retrieves the current number of file server hits. This is an administrative endpoint.
**Response:** HTML content displaying the number of visits.
**Status Codes:**
*   `200 OK`

---

## 3. POST /reset
**Description:** Resets the file server hit counter to 0, and also resets users, chirps, and refresh tokens. This is an administrative endpoint.
**Response:** `Server hits reset to: 0` (plain text)
**Status Codes:**
*   `200 OK`

---

## 4. GET /chirps
**Description:** Retrieves all chirps or chirps by a specific author.
**Query Parameters:**
*   `auther_id` (optional): A UUID to filter chirps by a specific user.
**Response (Success):** An array of Chirp objects.
```json
[
  {
    "id": "uuid",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "body": "string",
    "user_id": "uuid"
  }
]
```
**Response (Error):**
```json
{
  "error": "error message"
}
```
**Status Codes:**
*   `200 OK`
*   `404 Not Found` (if chirps or user not found)
*   `500 Internal Server Error` (for other errors, e.g., invalid `auther_id` format)

---

## 5. GET /chirps/{id}
**Description:** Retrieves a single chirp by its ID.
**URL Parameters:**
*   `id` (UUID): The ID of the chirp to retrieve.
**Response (Success):** A single Chirp object.
```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "body": "string",
  "user_id": "uuid"
}
```
**Response (Error):**
```json
{
  "error": "error message"
}
```
**Status Codes:**
*   `200 OK`
*   `404 Not Found` (if chirp not found)
*   `500 Internal Server Error` (for other errors)

---

## 6. POST /chirps
**Description:** Creates a new chirp. Requires authentication.
**Request Body:**
```json
{
  "body": "string"
}
```
**Headers:**
*   `Authorization`: `Bearer {JWT_TOKEN}`
**Response (Success):** The newly created Chirp object.
```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "body": "string",
  "user_id": "uuid"
}
```
**Response (Error):**
```json
{
  "error": "error message"
}
```
**Status Codes:**
*   `201 Created`
*   `400 Bad Request` (if chirp body is too long or invalid JSON)
*   `401 Unauthorized` (if JWT is missing or invalid)
*   `500 Internal Server Error` (for other errors)

---

## 7. DELETE /chirps/{id}
**Description:** Deletes a chirp by its ID. Requires authentication and ownership of the chirp.
**URL Parameters:**
*   `id` (UUID): The ID of the chirp to delete.
**Headers:**
*   `Authorization`: `Bearer {JWT_TOKEN}`
**Response:** Empty body on success.
**Response (Error):**
```json
{
  "error": "error message"
}
```
**Status Codes:**
*   `204 No Content`
*   `401 Unauthorized` (if JWT is missing or invalid)
*   `403 Forbidden` (if user does not own the chirp)
*   `404 Not Found` (if chirp not found)
*   `500 Internal Server Error` (for other errors)

---

## 8. POST /users
**Description:** Creates a new user.
**Request Body:**
```json
{
  "email": "string",
  "password": "string"
}
```
**Response (Success):** The newly created User object (without password).
```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "string",
  "is_chirpy_red": "boolean"
}
```
**Response (Error):**
```json
{
  "error": "error message"
}
```
**Status Codes:**
*   `201 Created`
*   `400 Bad Request` (if invalid JSON)
*   `500 Internal Server Error` (for other errors, e.g., email already exists)

---

## 9. PUT /users
**Description:** Updates an existing user's email and/or password. Requires authentication.
**Request Body:**
```json
{
  "email": "string",
  "password": "string"
}
```
**Headers:**
*   `Authorization`: `Bearer {JWT_TOKEN}`
**Response (Success):** The updated User object (without password).
```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "string",
  "is_chirpy_red": "boolean"
}
```
**Response (Error):**
```json
{
  "error": "error message"
}
```
**Status Codes:**
*   `200 OK`
*   `400 Bad Request` (if invalid JSON)
*   `401 Unauthorized` (if JWT is missing or invalid)
*   `500 Internal Server Error` (for other errors)

---

## 10. POST /login
**Description:** Authenticates a user and returns JWT and Refresh Tokens.
**Request Body:**
```json
{
  "email": "string",
  "password": "string",
  "expires_in_seconds": "integer (optional, default: 60)"
}
```
**Response (Success):** User data with JWT and Refresh Token.
```json
{
  "id": "uuid",
  "email": "string",
  "is_chirpy_red": "boolean",
  "token": "string (JWT)",
  "refresh_token": "string"
}
```
**Response (Error):**
```json
{
  "error": "error message"
}
```
**Status Codes:**
*   `200 OK`
*   `400 Bad Request` (if invalid JSON)
*   `401 Unauthorized` (if email/password is incorrect)
*   `500 Internal Server Error` (for other errors)

---

## 11. POST /refresh
**Description:** Refreshes an expired JWT using a valid refresh token.
**Headers:**
*   `Authorization`: `Bearer {REFRESH_TOKEN}`
**Response (Success):** New JWT and the same refresh token.
```json
{
  "token": "string (JWT)",
  "refresh_token": "string"
}
```
**Response (Error):**
```json
{
  "error": "error message"
}
```
**Status Codes:**
*   `200 OK`
*   `401 Unauthorized` (if refresh token is missing or invalid)
*   `500 Internal Server Error` (for other errors)

---

## 12. POST /revoke
**Description:** Revokes a refresh token, invalidating it for future use.
**Headers:**
*   `Authorization`: `Bearer {REFRESH_TOKEN}`
**Response:** Empty body on success.
**Response (Error):**
```json
{
  "error": "error message"
}
```
**Status Codes:**
*   `204 No Content`
*   `401 Unauthorized` (if refresh token is missing or invalid)
*   `500 Internal Server Error` (for other errors)
