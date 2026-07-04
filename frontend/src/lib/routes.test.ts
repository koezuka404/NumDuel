import { describe, expect, it } from 'vitest';
import { homePathForRole } from './routes';

describe('homePathForRole', () => {
  it('returns admin path for master', () => {
    expect(homePathForRole('master')).toBe('/admin');
  });

  it('returns matching path for user', () => {
    expect(homePathForRole('user')).toBe('/matching');
  });
});
