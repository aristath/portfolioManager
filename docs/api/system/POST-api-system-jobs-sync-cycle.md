# POST /api/system/jobs/sync-cycle

Trigger sync cycle job manually.

**Description:**
Manually triggers the sync cycle job to run immediately. The sync cycle job performs a complete synchronization cycle including syncing trades, cash flows, portfolio data, and prices from the broker. This endpoint enqueues the job in the queue system with high priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/sync-cycle`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Sync cycle triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Sync cycle job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a sync cycle job in the queue system with high priority
- Job executes asynchronously in the background
- Syncs trades, cash flows, portfolio, and prices from broker
- Updates local databases with latest broker data

---
