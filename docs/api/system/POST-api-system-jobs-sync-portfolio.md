# POST /api/system/jobs/sync-portfolio

Trigger Sync portfolio job manually.

**Description:**
Manually triggers the sync portfolio job to run immediately. The sync portfolio job synchronizes current portfolio positions and balances from the broker API to local databases. This endpoint enqueues the job in the queue system with high priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/sync-portfolio`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Sync portfolio triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Sync portfolio job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Sync portfolio job in the queue system with high priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
