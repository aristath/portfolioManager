# POST /api/system/jobs/dividend-reinvestment

Trigger dividend reinvestment job manually.

**Description:**
Manually triggers the dividend reinvestment (DRIP) job to run immediately. The dividend reinvestment job processes unreinvested dividends, creates dividend recommendations, and executes dividend reinvestment trades according to DRIP settings. This endpoint enqueues the job in the queue system with high priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/dividend-reinvestment`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Dividend reinvestment triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Dividend reinvestment job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a dividend reinvestment job in the queue system with high priority
- Job executes asynchronously in the background
- Processes unreinvested dividends and creates reinvestment recommendations
- May execute dividend reinvestment trades if DRIP is enabled

---
