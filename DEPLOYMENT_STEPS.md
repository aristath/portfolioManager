# Deployment Steps for Arduino Device

> **NOTE:** This guide is for the legacy monolith deployment (main branch).
> For microservices deployment (micro-services branch - RECOMMENDED), see [INSTALL.md](INSTALL.md).

## Steps to Apply New Python-Based Deployment System

### 1. SSH into Arduino
```bash
ssh arduino@192.168.1.11
# Password: aristath
```

### 2. Pull Latest Changes
```bash
cd /home/arduino/repos/autoTrader
git pull
```

### 3. Restart Service to Load New Code
```bash
sudo systemctl restart arduino-trader
sleep 3
sudo systemctl status arduino-trader
```

### 4. Remove Old Bash Script (No Longer Used)
```bash
rm -f /home/arduino/bin/auto-deploy.sh
ls -la /home/arduino/bin/ | grep auto-deploy || echo "Old script removed"
```

### 5. Check Deployment Status
```bash
# Check deployment status via API
curl http://localhost:8000/api/status/deploy/status

# Check job status
curl http://localhost:8000/api/status/jobs | grep -A 3 auto_deploy
```

### 6. Check Logs for Issues
```bash
# Check recent logs for deployment activity
sudo journalctl -u arduino-trader -n 100 | grep -i "deploy\|error\|exception"

# Follow logs in real-time
sudo journalctl -u arduino-trader -f
```

### 7. Test Manual Deployment Trigger
```bash
# Manually trigger deployment
curl -X POST http://localhost:8000/api/status/deploy/trigger

# Wait a few seconds, then check status
curl http://localhost:8000/api/status/deploy/status
```

## Verification

After applying changes, verify:
- ✅ Service is running: `sudo systemctl is-active arduino-trader`
- ✅ Deployment status API works: `curl http://localhost:8000/api/status/deploy/status`
- ✅ Old bash script is removed: `ls /home/arduino/bin/auto-deploy.sh` (should not exist)
- ✅ No errors in logs: `sudo journalctl -u arduino-trader -n 50 | grep -i error`

## Expected Behavior

The new system will:
- Check for updates every 5 minutes (configurable)
- Use Python-based deployment with retry logic
- Provide better error handling and logging
- Support deployment status API endpoints
- Use file-based locking to prevent concurrent deployments
