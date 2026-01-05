package deployment

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ServiceManager manages systemd services
type ServiceManager struct {
	log Logger
}

// NewServiceManager creates a new service manager
func NewServiceManager(log Logger) *ServiceManager {
	return &ServiceManager{
		log: log,
	}
}

// StopService stops a systemd service
// Tries multiple methods to work around NoNewPrivileges restriction
func (s *ServiceManager) StopService(serviceName string) error {
	s.log.Info().
		Str("service", serviceName).
		Msg("Stopping systemd service")

	// Method 1: Try systemctl directly (may work with polkit)
	cmd := exec.Command("systemctl", "stop", serviceName)
	_, err := cmd.CombinedOutput()
	if err == nil {
		s.log.Info().
			Str("service", serviceName).
			Msg("Successfully stopped systemd service (direct)")
		return nil
	}

	// Method 2: Try sudo (may fail with NoNewPrivileges, but worth trying)
	cmd = exec.Command("sudo", "systemctl", "stop", serviceName)
	_, err = cmd.CombinedOutput()
	if err == nil {
		s.log.Info().
			Str("service", serviceName).
			Msg("Successfully stopped systemd service (sudo)")
		return nil
	}

	// Method 3: Try using dbus-send (works without sudo if user has polkit permissions)
	// Convert service name to D-Bus object path format (e.g., "trader.service" -> "trader_2eservice")
	dbusPath := strings.ReplaceAll(serviceName, "-", "_2d")
	if !strings.HasSuffix(dbusPath, "_2eservice") {
		dbusPath = dbusPath + "_2eservice"
	}
	unitPath := fmt.Sprintf("/org/freedesktop/systemd1/unit/%s", dbusPath)
	cmd = exec.Command("dbus-send", "--system", "--print-reply",
		"--dest=org.freedesktop.systemd1",
		unitPath,
		"org.freedesktop.systemd1.Unit.Stop", "replace", "s", "")
	dbusOutput, err := cmd.CombinedOutput()
	if err == nil {
		s.log.Info().
			Str("service", serviceName).
			Msg("Successfully stopped systemd service (dbus)")
		return nil
	}

	// All methods failed - use last error output
	outputStr := strings.TrimSpace(string(dbusOutput))
	return &ServiceRestartError{
		ServiceName: serviceName,
		Message:     fmt.Sprintf("systemctl stop failed (NoNewPrivileges restriction): %s", outputStr),
		Err:         err,
	}
}

// StartService starts a systemd service
// Tries multiple methods to work around NoNewPrivileges restriction
func (s *ServiceManager) StartService(serviceName string) error {
	s.log.Info().
		Str("service", serviceName).
		Msg("Starting systemd service")

	// Method 1: Try systemctl directly (may work with polkit)
	cmd := exec.Command("systemctl", "start", serviceName)
	_, err := cmd.CombinedOutput()
	if err == nil {
		s.log.Info().
			Str("service", serviceName).
			Msg("Successfully started systemd service (direct)")
		return nil
	}

	// Method 2: Try sudo (may fail with NoNewPrivileges, but worth trying)
	cmd = exec.Command("sudo", "systemctl", "start", serviceName)
	_, err = cmd.CombinedOutput()
	if err == nil {
		s.log.Info().
			Str("service", serviceName).
			Msg("Successfully started systemd service (sudo)")
		return nil
	}

	// Method 3: Try using dbus-send (works without sudo if user has polkit permissions)
	// Convert service name to D-Bus object path format (e.g., "trader.service" -> "trader_2eservice")
	dbusPath := strings.ReplaceAll(serviceName, "-", "_2d")
	if !strings.HasSuffix(dbusPath, "_2eservice") {
		dbusPath = dbusPath + "_2eservice"
	}
	unitPath := fmt.Sprintf("/org/freedesktop/systemd1/unit/%s", dbusPath)
	cmd = exec.Command("dbus-send", "--system", "--print-reply",
		"--dest=org.freedesktop.systemd1",
		unitPath,
		"org.freedesktop.systemd1.Unit.Start", "replace", "s", "")
	dbusOutput, err := cmd.CombinedOutput()
	if err == nil {
		s.log.Info().
			Str("service", serviceName).
			Msg("Successfully started systemd service (dbus)")
		return nil
	}

	// All methods failed - use last error output
	outputStr := strings.TrimSpace(string(dbusOutput))
	return &ServiceRestartError{
		ServiceName: serviceName,
		Message:     fmt.Sprintf("systemctl start failed (NoNewPrivileges restriction): %s", outputStr),
		Err:         err,
	}
}

// RestartService restarts a systemd service
// Tries multiple methods to work around NoNewPrivileges restriction
func (s *ServiceManager) RestartService(serviceName string) error {
	s.log.Info().
		Str("service", serviceName).
		Msg("Restarting systemd service")

	// Method 1: Try systemctl directly (may work with polkit)
	cmd := exec.Command("systemctl", "restart", serviceName)
	_, err := cmd.CombinedOutput()
	if err == nil {
		s.log.Info().
			Str("service", serviceName).
			Msg("Successfully restarted systemd service (direct)")
		return nil
	}

	// Method 2: Try sudo (may fail with NoNewPrivileges, but worth trying)
	cmd = exec.Command("sudo", "systemctl", "restart", serviceName)
	_, err = cmd.CombinedOutput()
	if err == nil {
		s.log.Info().
			Str("service", serviceName).
			Msg("Successfully restarted systemd service (sudo)")
		return nil
	}

	// Method 3: Try using dbus-send (works without sudo if user has polkit permissions)
	// Convert service name to D-Bus object path format (e.g., "trader.service" -> "trader_2eservice")
	dbusPath := strings.ReplaceAll(serviceName, "-", "_2d")
	if !strings.HasSuffix(dbusPath, "_2eservice") {
		dbusPath = dbusPath + "_2eservice"
	}
	unitPath := fmt.Sprintf("/org/freedesktop/systemd1/unit/%s", dbusPath)
	cmd = exec.Command("dbus-send", "--system", "--print-reply",
		"--dest=org.freedesktop.systemd1",
		unitPath,
		"org.freedesktop.systemd1.Unit.Restart", "replace", "s", "")
	dbusOutput, err := cmd.CombinedOutput()
	if err == nil {
		s.log.Info().
			Str("service", serviceName).
			Msg("Successfully restarted systemd service (dbus)")
		return nil
	}

	// All methods failed - use last error output
	outputStr := strings.TrimSpace(string(dbusOutput))
	s.log.Warn().
		Str("service", serviceName).
		Str("error", outputStr).
		Msg("All restart methods failed, service may need manual restart")

	// Return error but don't fail deployment - binary was deployed successfully
	// The user can restart manually or fix the systemd configuration
	return &ServiceRestartError{
		ServiceName: serviceName,
		Message:     fmt.Sprintf("systemctl restart failed (NoNewPrivileges restriction): %s. Binary deployed successfully, but service restart requires manual intervention or systemd configuration change.", outputStr),
		Err:         err,
	}
}

// RestartServices restarts multiple services in parallel
func (s *ServiceManager) RestartServices(serviceNames []string) map[string]error {
	var wg sync.WaitGroup
	errors := make(map[string]error)
	var mu sync.Mutex

	for _, serviceName := range serviceNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			err := s.RestartService(name)
			if err != nil {
				mu.Lock()
				errors[name] = err
				mu.Unlock()
			}
		}(serviceName)
	}

	wg.Wait()
	return errors
}

// CheckHealth performs a health check on a service
func (s *ServiceManager) CheckHealth(apiURL string, maxAttempts int, timeout time.Duration) error {
	client := &http.Client{
		Timeout: timeout,
	}

	var lastError error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		s.log.Debug().
			Str("url", apiURL).
			Int("attempt", attempt).
			Int("max_attempts", maxAttempts).
			Msg("Performing health check")

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return &HealthCheckError{
				ServiceName: apiURL,
				Message:     "failed to create health check request",
				Err:         err,
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			lastError = err
			if attempt < maxAttempts {
				time.Sleep(1 * time.Second)
				continue
			}
			return &HealthCheckError{
				ServiceName: apiURL,
				Message:     fmt.Sprintf("health check failed after %d attempts", maxAttempts),
				Err:         err,
			}
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			s.log.Info().
				Str("url", apiURL).
				Int("status", resp.StatusCode).
				Msg("Health check passed")
			return nil
		}

		lastError = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		if attempt < maxAttempts {
			time.Sleep(1 * time.Second)
			continue
		}
	}

	return &HealthCheckError{
		ServiceName: apiURL,
		Message:     fmt.Sprintf("health check failed after %d attempts", maxAttempts),
		Err:         lastError,
	}
}

// GetServiceStatus returns the status of a systemd service
func (s *ServiceManager) GetServiceStatus(serviceName string) (string, error) {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get service status: %w", err)
	}

	status := strings.TrimSpace(string(output))
	return status, nil
}
