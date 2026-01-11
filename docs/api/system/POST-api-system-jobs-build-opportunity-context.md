# POST /api/system/jobs/build-opportunity-context

Trigger Build opportunity context job manually.

**Description:**
Manually triggers the build opportunity context job to run immediately. The build opportunity context job prepares the context data needed for opportunity identification and trade planning. This endpoint enqueues the job in the queue system with medium priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/build-opportunity-context`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Build opportunity context triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Build opportunity context job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Build opportunity context job in the queue system with medium priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
