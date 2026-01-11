# POST /api/system/jobs/check-negative-balances

Trigger Check negative balances job manually.

**Description:**
Manually triggers the check negative balances job to run immediately. The check negative balances job validates that no cash balances are negative and triggers alerts if any are found. This endpoint enqueues the job in the queue system with high priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/check-negative-balances`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Check negative balances triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Check negative balances job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Check negative balances job in the queue system with high priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
