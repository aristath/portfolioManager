# POST /api/system/jobs/execute-dividend-trades

Trigger Execute dividend trades job manually.

**Description:**
Manually triggers the execute dividend trades job to run immediately. The execute dividend trades job executes dividend reinvestment trades based on dividend recommendations. This endpoint enqueues the job in the queue system with high priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/execute-dividend-trades`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Execute dividend trades triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Execute dividend trades job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Execute dividend trades job in the queue system with high priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
