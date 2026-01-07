/**
 * Utility functions for date parsing and formatting
 */

/**
 * Parses a date string and returns a Date object or null if invalid
 * @param dateStr - ISO 8601 date string
 * @returns Date object or null if invalid
 */
export function parseDate(dateStr?: string): Date | null {
  if (!dateStr) return null;
  try {
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return null;
    return date;
  } catch {
    return null;
  }
}

/**
 * Formats a date string to a human-readable format
 * @param dateStr - ISO 8601 date string
 * @returns Formatted date string or 'N/A' if invalid
 */
export function formatDate(dateStr?: string): string {
  if (!dateStr) return 'N/A';
  const date = parseDate(dateStr);
  if (!date) return dateStr;
  
  try {
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return dateStr;
  }
}

