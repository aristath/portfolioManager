# POST /api/system/jobs/check-wal-checkpoints

Trigger Check WAL checkpoints job manually.

**Description:**
Manually triggers the check WAL checkpoints job to run immediately. The check WAL checkpoints job validates SQLite WAL (Write-Ahead Logging) checkpoint status and triggers checkpoints if needed to maintain database performance. This endpoint enqueues the job in the queue system with medium priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/check-wal-checkpoints`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Check WAL checkpoints triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Check WAL checkpoints job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Check WAL checkpoints job in the queue system with medium priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
