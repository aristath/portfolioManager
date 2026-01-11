# POST /api/system/jobs/check-core-databases

Trigger Check core databases job manually.

**Description:**
Manually triggers the check core databases job to run immediately. The check core databases job performs health checks on core databases (universe, config, ledger, portfolio) to ensure they are accessible and functioning correctly. This endpoint enqueues the job in the queue system with high priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/check-core-databases`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Check core databases triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Check core databases job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Check core databases job in the queue system with high priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
