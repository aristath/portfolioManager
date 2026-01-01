// Arduino Trader LED Display
// System stats visualization with random pixel density and microservice health
// Controls 8x13 LED matrix and RGB LEDs 3 & 4 on Arduino UNO Q
// Uses Router Bridge for communication with Linux MPU

#include <Arduino_RouterBridge.h>
#include "Arduino_LED_Matrix.h"

ArduinoLEDMatrix matrix;

// RGB LED pins use LED_BUILTIN offsets (from official unoq-pin-toggle example)
// LED3: LED_BUILTIN (R), LED_BUILTIN+1 (G), LED_BUILTIN+2 (B)
// LED4: LED_BUILTIN+3 (R), LED_BUILTIN+4 (G), LED_BUILTIN+5 (B)
// Active-low: HIGH = OFF, LOW = ON

// LED Matrix state
const int MATRIX_WIDTH = 13;
const int MATRIX_HEIGHT = 8;
const int TOTAL_PIXELS = MATRIX_WIDTH * MATRIX_HEIGHT;  // 104 pixels

// PWM brightness control (0-255 per pixel)
// Memory usage: 8 * 13 = 104 bytes
uint8_t pixelBrightness[MATRIX_HEIGHT][MATRIX_WIDTH];
float targetFillPercentage = 0.0;  // 0.0-100.0

// Two-layer timing: PWM frame cycling + pixel pattern updates
const int PWM_FRAME_INTERVAL = 10;     // Fixed 100 FPS for PWM rendering
unsigned long lastPWMFrame = 0;
uint8_t pwmCycle = 0;                   // 0-15 for 4-bit PWM

int pixelUpdateInterval = 2010;         // Dynamic interval for pattern updates
unsigned long lastPixelUpdate = 0;

// Shared frame buffer for rendering (8 * 12 = 96 bytes)
// Using global to avoid stack allocation in renderPWMFrame()
uint8_t frameBuffer[8][12];

// Set RGB LED 3 color (active-low, digital only)
void setRGB3(uint8_t r, uint8_t g, uint8_t b) {
  digitalWrite(LED_BUILTIN, r > 0 ? LOW : HIGH);
  digitalWrite(LED_BUILTIN + 1, g > 0 ? LOW : HIGH);
  digitalWrite(LED_BUILTIN + 2, b > 0 ? LOW : HIGH);
}

// Set RGB LED 4 color (active-low, digital only)
void setRGB4(uint8_t r, uint8_t g, uint8_t b) {
  digitalWrite(LED_BUILTIN + 3, r > 0 ? LOW : HIGH);
  digitalWrite(LED_BUILTIN + 4, g > 0 ? LOW : HIGH);
  digitalWrite(LED_BUILTIN + 5, b > 0 ? LOW : HIGH);
}

// Set fill percentage for LED matrix (0.0-100.0) with dynamic update interval
void setFillPercentageWithActivity(float percentage) {
  targetFillPercentage = constrain(percentage, 0.0, 100.0);

  // Calculate dynamic pixel update interval: 2010ms at 0% to 10ms at 100%
  pixelUpdateInterval = 2010 - (int)(targetFillPercentage * 20.0);
  pixelUpdateInterval = constrain(pixelUpdateInterval, 10, 2010);
}

// Update pixel pattern with random brightness based on activity level
void updatePixelPattern() {
  int targetLitPixels = (int)(targetFillPercentage / 100.0 * TOTAL_PIXELS);
  targetLitPixels = constrain(targetLitPixels, 0, TOTAL_PIXELS);

  // Calculate brightness range based on activity
  // At 0%: min=60, max=100
  // At 100%: min=215, max=255
  float activityPercent = constrain(targetFillPercentage, 0.0, 100.0);
  float minBrightness = 60.0 + (activityPercent * 1.55);
  float maxBrightness = minBrightness + 40.0;

  // Ensure brightness values are within valid range
  minBrightness = constrain(minBrightness, 16.0, 255.0);  // Min 16 for visibility
  maxBrightness = constrain(maxBrightness, minBrightness, 255.0);

  // Count currently lit pixels
  int currentLitPixels = 0;
  for (int y = 0; y < MATRIX_HEIGHT; y++) {
    for (int x = 0; x < MATRIX_WIDTH; x++) {
      if (pixelBrightness[y][x] > 0) {
        currentLitPixels++;
      }
    }
  }

  // Smoothly transition: add or remove pixels, and update brightness
  // This creates a more organic, flowing effect
  if (currentLitPixels < targetLitPixels) {
    // Need to add pixels
    int toAdd = targetLitPixels - currentLitPixels;
    int added = 0;
    int attempts = 0;
    const int MAX_ATTEMPTS = toAdd * 10;

    while (added < toAdd && attempts < MAX_ATTEMPTS) {
      int x = random(MATRIX_WIDTH);
      int y = random(MATRIX_HEIGHT);

      if (pixelBrightness[y][x] == 0) {
        pixelBrightness[y][x] = random((int)minBrightness, (int)maxBrightness + 1);
        added++;
      }
      attempts++;
    }
  } else if (currentLitPixels > targetLitPixels) {
    // Need to remove pixels
    int toRemove = currentLitPixels - targetLitPixels;
    int removed = 0;
    int attempts = 0;
    const int MAX_ATTEMPTS = toRemove * 10;

    while (removed < toRemove && attempts < MAX_ATTEMPTS) {
      int x = random(MATRIX_WIDTH);
      int y = random(MATRIX_HEIGHT);

      if (pixelBrightness[y][x] > 0) {
        pixelBrightness[y][x] = 0;
        removed++;
      }
      attempts++;
    }
  }

  // Add sparkle effect: randomly adjust brightness of some lit pixels
  // This creates visual interest even when pixel count is stable
  int sparkleCount = max(1, currentLitPixels / 10);  // 10% of lit pixels sparkle
  for (int i = 0; i < sparkleCount; i++) {
    int x = random(MATRIX_WIDTH);
    int y = random(MATRIX_HEIGHT);

    if (pixelBrightness[y][x] > 0) {
      // Adjust brightness within range
      pixelBrightness[y][x] = random((int)minBrightness, (int)maxBrightness + 1);
    }
  }
}

// Render PWM frame for software brightness control
void renderPWMFrame() {
  // Clear frame buffer
  memset(frameBuffer, 0, sizeof(frameBuffer));

  // Build frame based on current PWM cycle
  for (int y = 0; y < MATRIX_HEIGHT; y++) {
    for (int x = 0; x < MATRIX_WIDTH; x++) {
      // Get 4-bit brightness level (0-15)
      uint8_t level = pixelBrightness[y][x] >> 4;

      // Pixel is ON if pwmCycle < brightness level
      if (pwmCycle < level) {
        int byteIndex = x / 8;
        int bitIndex = 7 - (x % 8);
        if (byteIndex < 12) {  // Safety check
          frameBuffer[y][byteIndex] |= (1 << bitIndex);
        }
      }
    }
  }

  // Render to LED matrix
  matrix.renderBitmap(frameBuffer, 8, 12);

  // Increment PWM cycle (0-15 for 4-bit PWM)
  pwmCycle = (pwmCycle + 1) & 0x0F;
}

void setup() {
  // Initialize LED matrix
  matrix.begin();

  // Initialize pixel brightness array (all OFF)
  memset(pixelBrightness, 0, sizeof(pixelBrightness));

  // Initialize frame buffer (all OFF)
  memset(frameBuffer, 0, sizeof(frameBuffer));

  // Initialize RGB LED 3 & 4 pins
  pinMode(LED_BUILTIN, OUTPUT);
  pinMode(LED_BUILTIN + 1, OUTPUT);
  pinMode(LED_BUILTIN + 2, OUTPUT);
  pinMode(LED_BUILTIN + 3, OUTPUT);
  pinMode(LED_BUILTIN + 4, OUTPUT);
  pinMode(LED_BUILTIN + 5, OUTPUT);

  // Start with LEDs off (active-low: HIGH = OFF)
  setRGB3(0, 0, 0);
  setRGB4(0, 0, 0);

  // Setup Router Bridge
  Bridge.begin();
  Bridge.provide("setRGB3", setRGB3);
  Bridge.provide("setRGB4", setRGB4);
  Bridge.provide("setFillPercentageWithActivity", setFillPercentageWithActivity);

  // Seed random number generator
  randomSeed(analogRead(0));

  // Initial render
  renderPWMFrame();
}

void loop() {
  // Bridge handles RPC messages automatically in background thread
  // No need to call Bridge.loop() - it's handled by __loopHook()

  unsigned long currentMillis = millis();

  // Layer 1: PWM frame cycling at fixed 100 FPS
  if (currentMillis - lastPWMFrame >= PWM_FRAME_INTERVAL) {
    lastPWMFrame = currentMillis;
    renderPWMFrame();
  }

  // Layer 2: Pixel pattern updates at dynamic rate
  if (currentMillis - lastPixelUpdate >= pixelUpdateInterval) {
    lastPixelUpdate = currentMillis;
    updatePixelPattern();
  }

  // Small delay to allow Bridge background thread to process
  delay(1);
}
