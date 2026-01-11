# POST /api/system/jobs/tag-update

Trigger Tag update job manually.

**Description:**
Manually triggers the tag update job to run immediately. The tag update job updates security tags and metadata based on current market data and portfolio state. This endpoint enqueues the job in the queue system with medium priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/tag-update`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Tag update triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Tag update job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Tag update job in the queue system with medium priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
