/**
 * Sentinel LED Display - Arduino Uno Q
 * 
 * Controls the 8x13 LED matrix and RGB LEDs 3 & 4 on Arduino UNO Q.
 * Uses Arduino_RouterBridge for communication with the Linux MPU.
 * 
 * Following Arduino Uno Q documentation:
 * https://docs.arduino.cc/tutorials/uno-q/user-manual/
 * 
 * Hardware:
 * - LED Matrix: 8 rows x 13 columns (104 LEDs)
 * - RGB LED 3: LED3_R, LED3_G, LED3_B - Active LOW
 * - RGB LED 4: LED4_R, LED4_G, LED4_B - Active LOW
 * 
 * Bridge Functions Exposed:
 * - setRGB3(r, g, b) - Set RGB LED 3 color (sync indicator)
 * - setRGB4(r, g, b) - Set RGB LED 4 color (processing indicator)
 * - clearMatrix() - Clear the LED matrix
 * - setPixelCount(count) - Light up specific number of pixels
 * - drawMatrix(bytes) - Draw raw bitmap to matrix
 */

#include <Arduino_RouterBridge.h>
#include <Arduino_LED_Matrix.h>
#include <vector>

// LED Matrix instance
ArduinoLEDMatrix matrix;

// Matrix dimensions
const uint8_t MATRIX_ROWS = 8;
const uint8_t MATRIX_COLS = 13;

/**
 * Set RGB LED 3 color (sync indicator)
 * RGB LEDs are active-low: LOW = ON, HIGH = OFF
 * 
 * @param r Red value (0=OFF, >0=ON due to active-low)
 * @param g Green value
 * @param b Blue value
 */
void setRGB3(int r, int g, int b) {
    digitalWrite(LED3_R, r > 0 ? LOW : HIGH);
    digitalWrite(LED3_G, g > 0 ? LOW : HIGH);
    digitalWrite(LED3_B, b > 0 ? LOW : HIGH);
}

/**
 * Set RGB LED 4 color (processing indicator)
 * RGB LEDs are active-low: LOW = ON, HIGH = OFF
 */
void setRGB4(int r, int g, int b) {
    digitalWrite(LED4_R, r > 0 ? LOW : HIGH);
    digitalWrite(LED4_G, g > 0 ? LOW : HIGH);
    digitalWrite(LED4_B, b > 0 ? LOW : HIGH);
}

/**
 * Clear the LED matrix
 */
void clearMatrix() {
    matrix.clear();
}

/**
 * Set specific number of pixels to light up (for system stats visualization)
 * Fills pixels from left-to-right, top-to-bottom
 * 
 * @param pixelsOn Number of pixels to light (0-104)
 */
void setPixelCount(int pixelsOn) {
    pixelsOn = constrain(pixelsOn, 0, MATRIX_ROWS * MATRIX_COLS);
    
    uint8_t bitmap[MATRIX_ROWS][MATRIX_COLS];
    memset(bitmap, 0, sizeof(bitmap));
    
    int count = 0;
    for (int row = 0; row < MATRIX_ROWS && count < pixelsOn; row++) {
        for (int col = 0; col < MATRIX_COLS && count < pixelsOn; col++) {
            bitmap[row][col] = 1;
            count++;
        }
    }
    
    matrix.renderBitmap(bitmap, MATRIX_ROWS, MATRIX_COLS);
}

/**
 * Draw raw bitmap data to the matrix
 * Accepts a vector of bytes representing the 8x13 matrix state
 */
void drawMatrix(std::vector<uint8_t> data) {
    if (data.empty()) {
        return;
    }
    matrix.draw(data.data());
}

void setup() {
    // Initialize LED matrix
    matrix.begin();
    matrix.setGrayscaleBits(8);  // Support brightness levels
    matrix.clear();
    
    // Initialize RGB LED 3 pins (sync indicator)
    pinMode(LED3_R, OUTPUT);
    pinMode(LED3_G, OUTPUT);
    pinMode(LED3_B, OUTPUT);
    
    // Initialize RGB LED 4 pins (processing indicator)
    pinMode(LED4_R, OUTPUT);
    pinMode(LED4_G, OUTPUT);
    pinMode(LED4_B, OUTPUT);
    
    // Start with all LEDs off (active-low: HIGH = OFF)
    setRGB3(0, 0, 0);
    setRGB4(0, 0, 0);
    
    // Flash LED3 green briefly to indicate sketch started
    setRGB3(0, 255, 0);
    delay(500);
    setRGB3(0, 0, 0);
    
    // Initialize Bridge for communication with Linux MPU
    Bridge.begin();
    
    // Expose functions to Python via Bridge
    Bridge.provide("setRGB3", setRGB3);
    Bridge.provide("setRGB4", setRGB4);
    Bridge.provide("clearMatrix", clearMatrix);
    Bridge.provide("setPixelCount", setPixelCount);
    Bridge.provide("drawMatrix", drawMatrix);
}

void loop() {
    delay(10);
}
