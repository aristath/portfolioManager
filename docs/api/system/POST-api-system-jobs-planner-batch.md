# POST /api/system/jobs/planner-batch

Trigger Planner batch job manually.

**Description:**
Manually triggers the planner batch job to run immediately. The planner batch job generates trade plans and recommendations based on current portfolio state, allocation targets, and market conditions. This endpoint enqueues the job in the queue system with high priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/planner-batch`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Planner batch triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Planner batch job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Planner batch job in the queue system with high priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
