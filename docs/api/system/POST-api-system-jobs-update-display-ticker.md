# POST /api/system/jobs/update-display-ticker

Trigger Update display ticker job manually.

**Description:**
Manually triggers the update display ticker job to run immediately. The update display ticker job refreshes the LED display ticker content with current portfolio information. This endpoint enqueues the job in the queue system with medium priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/update-display-ticker`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Update display ticker triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Update display ticker job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Update display ticker job in the queue system with medium priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
