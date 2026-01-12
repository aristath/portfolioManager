package mcu

// ScrollText scrolls text across the LED matrix.
// text: The text to scroll
// speed: Milliseconds per scroll step (lower = faster)
func (c *Client) ScrollText(text string, speed int) error {
	_, err := c.Call("scrollText", text, speed)
	return err
}

// SetRGB3 sets the color of RGB LED 3 (sync indicator).
// r, g, b: Color values 0-255 (>0 = on for that channel)
func (c *Client) SetRGB3(r, g, b int) error {
	_, err := c.Call("setRGB3", r, g, b)
	return err
}

// SetRGB4 sets the color of RGB LED 4 (processing indicator).
// r, g, b: Color values 0-255 (>0 = on for that channel)
func (c *Client) SetRGB4(r, g, b int) error {
	_, err := c.Call("setRGB4", r, g, b)
	return err
}

// SetBlink3 enables blink mode for LED3.
// r, g, b: Color values 0-255
// intervalMs: Blink interval in milliseconds
func (c *Client) SetBlink3(r, g, b, intervalMs int) error {
	_, err := c.Call("setBlink3", r, g, b, intervalMs)
	return err
}

// StopBlink3 stops LED3 blinking and turns it off.
func (c *Client) StopBlink3() error {
	_, err := c.Call("stopBlink3")
	return err
}

// SetBlink4 enables simple blink mode for LED4.
// r, g, b: Color values 0-255
// intervalMs: Blink interval in milliseconds
func (c *Client) SetBlink4(r, g, b, intervalMs int) error {
	_, err := c.Call("setBlink4", r, g, b, intervalMs)
	return err
}

// SetBlink4Alternating enables alternating color mode for LED4.
// r1, g1, b1: First color RGB values 0-255
// r2, g2, b2: Second color RGB values 0-255
// intervalMs: Blink interval in milliseconds
func (c *Client) SetBlink4Alternating(r1, g1, b1, r2, g2, b2, intervalMs int) error {
	_, err := c.Call("setBlink4Alternating", r1, g1, b1, r2, g2, b2, intervalMs)
	return err
}

// SetBlink4Coordinated enables coordinated mode for LED4 (alternates with LED3).
// r, g, b: Color values 0-255
// intervalMs: Blink interval in milliseconds
// led3Phase: LED3 phase state (true = LED3 is on)
func (c *Client) SetBlink4Coordinated(r, g, b, intervalMs int, led3Phase bool) error {
	_, err := c.Call("setBlink4Coordinated", r, g, b, intervalMs, led3Phase)
	return err
}

// StopBlink4 stops LED4 blinking and turns it off.
func (c *Client) StopBlink4() error {
	_, err := c.Call("stopBlink4")
	return err
}

// SetSystemStats sets the system stats visualization mode on the LED matrix.
// pixelsOn: Number of pixels to light up (0-104)
// brightness: Brightness level (100-220)
// intervalMs: Animation interval in milliseconds
func (c *Client) SetSystemStats(pixelsOn, brightness, intervalMs int) error {
	_, err := c.Call("setSystemStats", pixelsOn, brightness, intervalMs)
	return err
}

// SetPortfolioMode sets the portfolio visualization mode on the LED matrix.
// clustersJSON: JSON string containing cluster data
func (c *Client) SetPortfolioMode(clustersJSON string) error {
	_, err := c.Call("setPortfolioMode", clustersJSON)
	return err
}
