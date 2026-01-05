/**
 * Maps tag IDs to human-readable names
 * Based on TAG_SUGGESTIONS.md
 */

const TAG_NAMES = {
  // Opportunity Tags - Value
  'value-opportunity': 'Value Opportunity',
  'deep-value': 'Deep Value',
  'below-52w-high': 'Below 52-Week High',
  'undervalued-pe': 'Undervalued P/E',

  // Opportunity Tags - Quality
  'high-quality': 'High Quality',
  'stable': 'Stable',
  'consistent-grower': 'Consistent Grower',
  'strong-fundamentals': 'Strong Fundamentals',

  // Opportunity Tags - Technical
  'oversold': 'Oversold',
  'below-ema': 'Below 200-Day EMA',
  'bollinger-oversold': 'Near Bollinger Lower Band',

  // Opportunity Tags - Dividend
  'high-dividend': 'High Dividend Yield',
  'dividend-opportunity': 'Dividend Opportunity',
  'dividend-grower': 'Dividend Grower',

  // Opportunity Tags - Momentum
  'positive-momentum': 'Positive Momentum',
  'recovery-candidate': 'Recovery Candidate',

  // Opportunity Tags - Score
  'high-score': 'High Overall Score',
  'good-opportunity': 'Good Opportunity',

  // Danger Tags - Volatility
  'volatile': 'Volatile',
  'volatility-spike': 'Volatility Spike',
  'high-volatility': 'High Volatility',

  // Danger Tags - Overvaluation
  'overvalued': 'Overvalued',
  'near-52w-high': 'Near 52-Week High',
  'above-ema': 'Above 200-Day EMA',
  'overbought': 'Overbought',

  // Danger Tags - Instability
  'instability-warning': 'Instability Warning',
  'unsustainable-gains': 'Unsustainable Gains',
  'valuation-stretch': 'Valuation Stretch',

  // Danger Tags - Underperformance
  'underperforming': 'Underperforming',
  'stagnant': 'Stagnant',
  'high-drawdown': 'High Drawdown',

  // Danger Tags - Portfolio Risk
  'overweight': 'Overweight in Portfolio',
  'concentration-risk': 'Concentration Risk',

  // Characteristic Tags - Risk Profile
  'low-risk': 'Low Risk',
  'medium-risk': 'Medium Risk',
  'high-risk': 'High Risk',

  // Characteristic Tags - Growth Profile
  'growth': 'Growth',
  'value': 'Value',
  'dividend-focused': 'Dividend Focused',

  // Characteristic Tags - Time Horizon
  'long-term': 'Long-Term Promise',
  'short-term-opportunity': 'Short-Term Opportunity',
};

/**
 * Get human-readable name for a tag ID
 * @param {string} tagId - The tag ID (e.g., 'value-opportunity')
 * @returns {string} Human-readable name (e.g., 'Value Opportunity')
 */
export function getTagName(tagId) {
  return TAG_NAMES[tagId] || tagId;
}

/**
 * Get all tag names for an array of tag IDs
 * @param {string[]} tagIds - Array of tag IDs
 * @returns {string[]} Array of human-readable names
 */
export function getTagNames(tagIds) {
  if (!tagIds || !Array.isArray(tagIds)) return [];
  return tagIds.map(getTagName);
}

/**
 * Get tag color variant based on tag category
 * @param {string} tagId - The tag ID
 * @returns {object} Mantine Badge props with color and variant
 */
export function getTagColor(tagId) {
  // Opportunity tags - green/blue
  if (tagId.startsWith('value-') ||
      tagId.startsWith('high-') ||
      tagId.startsWith('good-') ||
      tagId === 'stable' ||
      tagId === 'oversold' ||
      tagId === 'positive-momentum' ||
      tagId === 'recovery-candidate' ||
      tagId === 'dividend-opportunity' ||
      tagId === 'dividend-grower' ||
      tagId === 'consistent-grower' ||
      tagId === 'strong-fundamentals' ||
      tagId === 'below-52w-high' ||
      tagId === 'below-ema' ||
      tagId === 'bollinger-oversold' ||
      tagId === 'undervalued-pe' ||
      tagId === 'deep-value' ||
      tagId === 'high-dividend' ||
      tagId === 'high-score' ||
      tagId === 'good-opportunity') {
    return { color: 'green', variant: 'light' };
  }

  // Danger tags - red/orange
  if (tagId.startsWith('volatile') ||
      tagId.startsWith('over') ||
      tagId.startsWith('under') ||
      tagId === 'stagnant' ||
      tagId === 'high-drawdown' ||
      tagId === 'instability-warning' ||
      tagId === 'unsustainable-gains' ||
      tagId === 'valuation-stretch' ||
      tagId === 'concentration-risk' ||
      tagId === 'overweight') {
    return { color: 'red', variant: 'light' };
  }

  // Characteristic tags - blue/gray
  if (tagId.startsWith('low-') ||
      tagId.startsWith('medium-') ||
      tagId.startsWith('high-risk') ||
      tagId === 'growth' ||
      tagId === 'value' ||
      tagId === 'dividend-focused' ||
      tagId === 'long-term' ||
      tagId === 'short-term-opportunity') {
    return { color: 'blue', variant: 'light' };
  }

  // Default
  return { color: 'gray', variant: 'light' };
}
