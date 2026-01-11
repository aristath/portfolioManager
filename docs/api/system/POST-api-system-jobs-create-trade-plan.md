# POST /api/system/jobs/create-trade-plan

Trigger Create trade plan job manually.

**Description:**
Manually triggers the create trade plan job to run immediately. The create trade plan job generates a trade plan based on current portfolio state, allocation targets, and opportunities. This endpoint enqueues the job in the queue system with high priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/create-trade-plan`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Create trade plan triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Create trade plan job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Create trade plan job in the queue system with high priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
