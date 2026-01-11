# POST /api/system/jobs/set-pending-bonuses

Trigger Set pending bonuses job manually.

**Description:**
Manually triggers the set pending bonuses job to run immediately. The set pending bonuses job processes pending bonus shares from dividend reinvestments and updates dividend records. This endpoint enqueues the job in the queue system with medium priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/set-pending-bonuses`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Set pending bonuses triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Set pending bonuses job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Set pending bonuses job in the queue system with medium priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
