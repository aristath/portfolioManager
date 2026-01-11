# POST /api/settings/restart-service

Restart the Sentinel service.

**Description:**
Restarts the Sentinel service using systemctl. This endpoint sends a restart command to the system service manager to restart the Sentinel application. The service will be stopped and then started again, which may cause a brief interruption in API availability.

**Request:**
- Method: `POST`
- Path: `/api/settings/restart-service`
- Body: None (empty request body)

**Response:**
- Status: `200 OK`
- Body:
  ```json
  {
    "status": "success",
    "message": "Service restart initiated"
  }
  ```
  - `status` (string): Response status ("success" or "error")
  - `message` (string): Human-readable message

**Error Responses:**
- `500 Internal Server Error`: Failed to restart service (e.g., systemctl command failed, insufficient permissions)

**Side Effects:**
- Executes `systemctl restart sentinel` command
- Service is stopped and then started
- Brief interruption in API availability during restart
- All in-memory state is lost
- Active connections are terminated

**Notes:**
- Requires system-level permissions to execute systemctl commands
- The restart is performed asynchronously - the response may return before the service is fully restarted
- All active API requests will be interrupted

---
