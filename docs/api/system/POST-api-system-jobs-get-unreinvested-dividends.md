# POST /api/system/jobs/get-unreinvested-dividends

Trigger Get unreinvested dividends job manually.

**Description:**
Manually triggers the get unreinvested dividends job to run immediately. The get unreinvested dividends job retrieves all dividends that have not yet been reinvested from the ledger database. This endpoint enqueues the job in the queue system with medium priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/get-unreinvested-dividends`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Get unreinvested dividends triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Get unreinvested dividends job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues a Get unreinvested dividends job in the queue system with medium priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
