import { createTheme } from '@mantine/core';

// Catppuccin Mocha color palette
// https://catppuccin.com/palette
const catppuccinMocha = {
  // Base colors
  base: '#1e1e2e',      // Base background
  mantle: '#181825',     // Secondary background
  crust: '#11111b',      // Tertiary background

  // Surface colors (lighter to darker)
  surface0: '#313244',   // Surface 0
  surface1: '#45475a',   // Surface 1
  surface2: '#585b70',   // Surface 2

  // Text colors
  text: '#cdd6f4',       // Primary text
  subtext1: '#bac2de',   // Secondary text
  subtext0: '#a6adc8',   // Tertiary text

  // Overlay colors
  overlay0: '#6c7086',
  overlay1: '#7f849c',
  overlay2: '#9399b2',

  // Accent colors
  blue: '#89b4fa',
  green: '#a6e3a1',
  red: '#f38ba8',
  yellow: '#f9e2af',
  peach: '#fab387',
  mauve: '#cba6f7',
  teal: '#94e2d5',
  sky: '#89dceb',
  sapphire: '#74c7ec',
  lavender: '#b4befe',
  pink: '#f5c2e7',
  rosewater: '#f5e0dc',
  flamingo: '#f2cdcd',
  maroon: '#eba0ac',
};

export const theme = createTheme({
  primaryColor: 'blue',
  defaultRadius: 'sm', // Smaller radius for terminal aesthetic
  fontFamily: '"JetBrains Mono", "Fira Code", "IBM Plex Mono", "Source Code Pro", "Consolas", "Monaco", "Courier New", monospace',
  fontSizes: {
    xs: '0.875rem',   // 14px
    sm: '1rem',       // 16px
    md: '1.125rem',   // 18px
    lg: '1.25rem',    // 20px
    xl: '1.5rem',     // 24px
  },
  headings: {
    fontFamily: '"JetBrains Mono", "Fira Code", "IBM Plex Mono", "Source Code Pro", "Consolas", "Monaco", "Courier New", monospace',
  },
  defaultGradient: {
    from: catppuccinMocha.blue,
    to: catppuccinMocha.mauve,
    deg: 45,
  },
  colors: {
    // Dark theme colors mapped from Catppuccin Mocha (darker variant)
    dark: [
      catppuccinMocha.text,        // 0 - lightest text
      catppuccinMocha.subtext1,    // 1
      catppuccinMocha.subtext0,    // 2
      catppuccinMocha.overlay2,    // 3
      catppuccinMocha.overlay1,    // 4
      catppuccinMocha.overlay0,    // 5
      catppuccinMocha.surface0,    // 6 - borders
      catppuccinMocha.base,        // 7 - panels/cards
      catppuccinMocha.mantle,      // 8 - secondary background
      catppuccinMocha.crust,       // 9 - darkest background (main)
    ],
    // Blue (primary)
    blue: [
      catppuccinMocha.blue,
      catppuccinMocha.sapphire,
      catppuccinMocha.sky,
      catppuccinMocha.blue,
      catppuccinMocha.blue,
      catppuccinMocha.blue,
      catppuccinMocha.blue,
      catppuccinMocha.blue,
      catppuccinMocha.blue,
      catppuccinMocha.blue,
    ],
    // Green (success)
    green: [
      catppuccinMocha.green,
      catppuccinMocha.green,
      catppuccinMocha.green,
      catppuccinMocha.green,
      catppuccinMocha.green,
      catppuccinMocha.green,
      catppuccinMocha.green,
      catppuccinMocha.green,
      catppuccinMocha.green,
      catppuccinMocha.green,
    ],
    // Red (error/danger)
    red: [
      catppuccinMocha.red,
      catppuccinMocha.maroon,
      catppuccinMocha.red,
      catppuccinMocha.red,
      catppuccinMocha.red,
      catppuccinMocha.red,
      catppuccinMocha.red,
      catppuccinMocha.red,
      catppuccinMocha.red,
      catppuccinMocha.red,
    ],
    // Yellow (warning)
    yellow: [
      catppuccinMocha.yellow,
      catppuccinMocha.peach,
      catppuccinMocha.yellow,
      catppuccinMocha.yellow,
      catppuccinMocha.yellow,
      catppuccinMocha.yellow,
      catppuccinMocha.yellow,
      catppuccinMocha.yellow,
      catppuccinMocha.yellow,
      catppuccinMocha.yellow,
    ],
    // Gray (neutral)
    gray: [
      catppuccinMocha.text,
      catppuccinMocha.subtext1,
      catppuccinMocha.subtext0,
      catppuccinMocha.overlay2,
      catppuccinMocha.overlay1,
      catppuccinMocha.overlay0,
      catppuccinMocha.surface2,
      catppuccinMocha.surface1,
      catppuccinMocha.surface0,
      catppuccinMocha.base,
    ],
    // Teal (info)
    teal: [
      catppuccinMocha.teal,
      catppuccinMocha.teal,
      catppuccinMocha.teal,
      catppuccinMocha.teal,
      catppuccinMocha.teal,
      catppuccinMocha.teal,
      catppuccinMocha.teal,
      catppuccinMocha.teal,
      catppuccinMocha.teal,
      catppuccinMocha.teal,
    ],
    // Violet (special)
    violet: [
      catppuccinMocha.mauve,
      catppuccinMocha.lavender,
      catppuccinMocha.mauve,
      catppuccinMocha.mauve,
      catppuccinMocha.mauve,
      catppuccinMocha.mauve,
      catppuccinMocha.mauve,
      catppuccinMocha.mauve,
      catppuccinMocha.mauve,
      catppuccinMocha.mauve,
    ],
  },
  other: {
    // Custom Catppuccin colors for direct access
    catppuccin: catppuccinMocha,
  },
});
