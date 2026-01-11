# POST /api/system/jobs/event-based-trading

Trigger Event-based trading job manually.

**Description:**
Manually triggers the event-based trading job to run immediately. The event-based trading job processes trading events and executes trades based on configured event triggers. This endpoint enqueues the job in the queue system with critical priority for asynchronous execution.

**Request:**
- Method: `POST`
- Path: `/api/system/jobs/event-based-trading`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Event-based trading triggered successfully"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `200 OK` (with error status): Job not registered yet
  ```json
  {
    "status": "error",
    "message": "Event-based trading job not registered"
  }
  ```
- `500 Internal Server Error`: Failed to enqueue job

**Side Effects:**
- Enqueues an event-based trading job in the queue system with critical priority
- Job executes asynchronously in the background
- Job-specific side effects vary by job type

---
