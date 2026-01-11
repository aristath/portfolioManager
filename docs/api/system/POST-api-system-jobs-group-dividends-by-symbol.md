# POST /api/system/jobs/group-dividends-by-symbol

Trigger Group dividends by symbol job manually.

**Description:**
Manually triggers the group dividends by symbol job to run immediately. The group dividends by symbol job aggregates dividends by security symbol for dividend reinvestment processing. This endpoint enqueues the job in the queue system with medium priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/group-dividends-by-symbol`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Group dividends by symbol triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Group dividends by symbol job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Group dividends by symbol job in the queue system with medium priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
