# POST /api/system/jobs/sync-cash-flows

Trigger Sync cash flows job manually.

**Description:**
Manually triggers the sync cash flows job to run immediately. The sync cash flows job synchronizes cash flow transactions (deposits, withdrawals, fees) from the broker API to local databases. This endpoint enqueues the job in the queue system with high priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/sync-cash-flows`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Sync cash flows triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Sync cash flows job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Sync cash flows job in the queue system with high priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
