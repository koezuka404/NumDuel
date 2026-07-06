import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { MockWebSocket } from '../test/mockWebSocket';
import { WebSocketProvider, useWebSocket } from './useWebSocket';

const useAuth = vi.fn();

vi.mock('./useAuth', () => ({
  useAuth: () => useAuth(),
}));

function wrapper({ children }: { children: React.ReactNode }) {
  return <WebSocketProvider>{children}</WebSocketProvider>;
}

function mockWsTicketFetch() {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ data: { ticket: 'test-ticket' } }),
    }),
  );
}

describe('useWebSocket', () => {
  beforeEach(() => {
    MockWebSocket.reset();
    vi.stubGlobal('WebSocket', MockWebSocket as unknown as typeof WebSocket);
    mockWsTicketFetch();
    useAuth.mockReturnValue({
      isAuthenticated: true,
      user: { id: '1', username: 'alice', role: 'user' },
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  it('connects, authenticates, sends, and subscribes', async () => {
    const listener = vi.fn();
    const { result } = renderHook(() => useWebSocket(), { wrapper });

    await waitFor(() => expect(MockWebSocket.instances.length).toBe(1));
    const ws = MockWebSocket.latest();
    act(() => {
      ws.simulateOpen();
      ws.simulateMessage({ type: 'AUTH_OK' });
    });
    await waitFor(() => expect(result.current.connected).toBe(true));

    result.current.subscribe(listener);
    act(() => {
      ws.simulateMessage({ type: 'PING', data: {} });
      result.current.send({ type: 'TEST' });
    });
    expect(listener).toHaveBeenCalled();
    expect(ws.sent.some((payload) => payload.includes('TEST'))).toBe(true);
  });

  it('handles token_expired and unauthorized errors', async () => {
    const fetchMock = vi.fn().mockResolvedValue({});
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() => useWebSocket(), { wrapper });
    await waitFor(() => expect(MockWebSocket.instances.length).toBe(1));
    const ws = MockWebSocket.latest();

    act(() => {
      ws.simulateOpen();
      ws.simulateMessage({ type: 'ERROR', data: { code: 'token_expired' } });
    });
    await waitFor(() => expect(fetchMock).toHaveBeenCalled());

    act(() => {
      ws.simulateMessage({ type: 'ERROR', data: { code: 'unauthorized' } });
    });
    await waitFor(() => expect(result.current.connected).toBe(false));
  });

  it('emits RECONNECT_FAILED after max reconnect attempts', async () => {
    vi.useFakeTimers();
    const listener = vi.fn();
    const { result } = renderHook(() => useWebSocket(), { wrapper });
    result.current.subscribe(listener);

    await waitFor(() => expect(MockWebSocket.instances.length).toBe(1));

    for (let attempt = 0; attempt < 6; attempt += 1) {
      const ws = MockWebSocket.latest();
      act(() => {
        ws.close();
      });
      if (attempt < 5) {
        act(() => {
          vi.advanceTimersByTime(1000 * 2 ** attempt);
        });
        await waitFor(() => expect(MockWebSocket.instances.length).toBeGreaterThan(attempt));
      }
    }

    expect(listener.mock.calls.some(([msg]) => msg.type === 'RECONNECT_FAILED')).toBe(true);
  });

  it('disconnects for non-user roles', async () => {
    useAuth.mockReturnValue({
      isAuthenticated: true,
      user: { id: '1', username: 'admin', role: 'master' },
    });
    renderHook(() => useWebSocket(), { wrapper });
    await waitFor(() => expect(MockWebSocket.instances.length).toBe(0));
  });

  it('throws outside provider', () => {
    expect(() => renderHook(() => useWebSocket())).toThrow(
      'useWebSocket must be used within WebSocketProvider',
    );
  });
});
