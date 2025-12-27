// Arduino Trader LED Display
// Simple text scroller for 8x13 LED matrix on Arduino UNO Q
// Uses Router Bridge for communication with Linux MPU

#include <Arduino_RouterBridge.h>
#include "ArduinoGraphics.h"
#include "Arduino_LED_Matrix.h"

ArduinoLEDMatrix matrix;

// Scroll text across LED matrix using native ArduinoGraphics
// text: String to scroll, speed: ms per scroll step (lower = faster)
void scrollText(String text, int speed) {
  matrix.textScrollSpeed(speed);
  matrix.textFont(Font_5x7);
  // Convert speed to brightness (use default brightness)
  uint32_t color = 0xFFFFFF;  // White (full brightness)
  matrix.beginText(13, 1, color);  // Start at X=13 (matrix width) to scroll in from right
  matrix.print(text);
  matrix.endText(SCROLL_LEFT);
}

void setup() {
  // Initialize LED matrix
  matrix.begin();
  matrix.textFont(Font_5x7);
  matrix.setGrayscaleBits(8);  // For 0-255 brightness values
  matrix.clear();

  // Setup Router Bridge
  Bridge.begin();
  Bridge.provide("scrollText", scrollText);
}

void loop() {
  delay(100);
}
