import { describe, expect, it } from 'vitest';
import { formatDateTime, truncateId } from './format';

describe('formatDateTime', () => {
  it('formats ISO date in ja-JP locale', () => {
    const formatted = formatDateTime('2024-01-15T12:00:00.000Z');
    expect(formatted).toContain('2024');
  });
});

describe('truncateId', () => {
  it('truncates with default length', () => {
    expect(truncateId('abcdefghijklmnop')).toBe('abcdefgh…');
  });

  it('truncates with custom length', () => {
    expect(truncateId('abcdefghijklmnop', 4)).toBe('abcd…');
  });
});
