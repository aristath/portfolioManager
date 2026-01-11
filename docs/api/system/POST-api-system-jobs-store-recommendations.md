# POST /api/system/jobs/store-recommendations

Trigger Store recommendations job manually.

**Description:**
Manually triggers the store recommendations job to run immediately. The store recommendations job saves generated trade recommendations to the cache database for retrieval by the frontend. This endpoint enqueues the job in the queue system with medium priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/store-recommendations`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Store recommendations triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Store recommendations job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Store recommendations job in the queue system with medium priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
