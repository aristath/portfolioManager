# POST /api/system/jobs/check-dividend-yields

Trigger Check dividend yields job manually.

**Description:**
Manually triggers the check dividend yields job to run immediately. The check dividend yields job calculates and validates dividend yields for securities in the portfolio. This endpoint enqueues the job in the queue system with medium priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/check-dividend-yields`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Check dividend yields triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Check dividend yields job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Check dividend yields job in the queue system with medium priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
