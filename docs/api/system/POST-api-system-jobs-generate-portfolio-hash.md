# POST /api/system/jobs/generate-portfolio-hash

Trigger Generate portfolio hash job manually.

**Description:**
Manually triggers the generate portfolio hash job to run immediately. The generate portfolio hash job calculates a hash of the current portfolio state for change detection and tracking. This endpoint enqueues the job in the queue system with medium priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/generate-portfolio-hash`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Generate portfolio hash triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Generate portfolio hash job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Generate portfolio hash job in the queue system with medium priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
