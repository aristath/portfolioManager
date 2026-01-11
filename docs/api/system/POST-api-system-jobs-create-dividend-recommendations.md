# POST /api/system/jobs/create-dividend-recommendations

Trigger Create dividend recommendations job manually.

**Description:**
Manually triggers the create dividend recommendations job to run immediately. The create dividend recommendations job generates dividend reinvestment recommendations based on unreinvested dividends and DRIP settings. This endpoint enqueues the job in the queue system with medium priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/create-dividend-recommendations`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Create dividend recommendations triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Create dividend recommendations job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Create dividend recommendations job in the queue system with medium priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
