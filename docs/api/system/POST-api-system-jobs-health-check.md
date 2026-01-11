# POST /api/system/jobs/health-check

Trigger health check job manually.

**Description:**
Manually triggers the health check job to run immediately. The health check job performs system health checks including database connectivity, disk space, and system resource monitoring. This endpoint enqueues the job in the queue system for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/health-check`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Health check triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Health check job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a health check job in the queue system
- Job executes asynchronously in the background
- Health check results are logged and may trigger alerts

---
