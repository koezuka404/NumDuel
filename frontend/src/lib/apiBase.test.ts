import { describe, expect, it, vi } from 'vitest';
import { resolveApiBaseURL, resolveWsBaseURL } from './apiBase';

describe('resolveApiBaseURL', () => {
  it('uses VITE_API_BASE_URL when set', () => {
    vi.stubEnv('VITE_API_BASE_URL', 'https://api.example.com');
    expect(resolveApiBaseURL()).toBe('https://api.example.com/api');
    vi.unstubAllEnvs();
  });

  it('keeps /api suffix when already present', () => {
    vi.stubEnv('VITE_API_BASE_URL', 'https://api.example.com/api/');
    expect(resolveApiBaseURL()).toBe('https://api.example.com/api');
    vi.unstubAllEnvs();
  });

  it('falls back to relative /api', () => {
    vi.unstubAllEnvs();
    expect(resolveApiBaseURL()).toBe('/api');
  });
});

describe('resolveWsBaseURL', () => {
  it('uses VITE_WS_BASE_URL when set', () => {
    vi.stubEnv('VITE_WS_BASE_URL', 'wss://api.example.com');
    expect(resolveWsBaseURL()).toBe('wss://api.example.com/ws');
    vi.unstubAllEnvs();
  });

  it('derives ws url from window location', () => {
    vi.unstubAllEnvs();
    const url = resolveWsBaseURL();
    expect(url.endsWith('/ws')).toBe(true);
  });
});
