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

describe('useWebSocket', () => {
  beforeEach(() => {
    MockWebSocket.reset();
    vi.stubGlobal('WebSocket', MockWebSocket as unknown as typeof WebSocket);
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
    });
    expect(ws.sent[0]).toContain('AUTH');

    act(() => {
      ws.simulateMessage({ type: 'AUTH_OK' });
    });
    await waitFor(() => expect(result.current.connected).toBe(true));

    const unsubscribe = result.current.subscribe(listener);
    act(() => {
      ws.simulateMessage({ type: 'PING', data: {} });
    });
    expect(listener).toHaveBeenCalled();
    unsubscribe();

    act(() => {
      result.current.send({ type: 'TEST' });
    });
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
    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith('/api/auth/refresh', expect.any(Object)));

    act(() => {
      ws.simulateMessage({ type: 'ERROR', data: { code: 'unauthorized' } });
    });
    await waitFor(() => expect(result.current.connected).toBe(false));
  });

  it('reconnects and emits RECONNECT_FAILED after max attempts', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const listener = vi.fn();
    const { result } = renderHook(() => useWebSocket(), { wrapper });
    result.current.subscribe(listener);

    await waitFor(() => expect(MockWebSocket.instances.length).toBe(1));
    const ws = MockWebSocket.latest();
    act(() => {
      ws.simulateOpen();
    });

    for (let attempt = 0; attempt < 6; attempt += 1) {
      act(() => {
        ws.close();
      });
      await act(async () => {
        await vi.advanceTimersByTimeAsync(60_000);
      });
    }

    await waitFor(() =>
      expect(listener.mock.calls.some(([msg]) => msg.type === 'RECONNECT_FAILED')).toBe(true),
    );
  });

  it('disconnects for non-user roles and ignores invalid messages', async () => {
    useAuth.mockReturnValue({
      isAuthenticated: true,
      user: { id: '1', username: 'admin', role: 'master' },
    });
    renderHook(() => useWebSocket(), { wrapper });
    await waitFor(() => expect(MockWebSocket.instances.length).toBe(0));

    useAuth.mockReturnValue({
      isAuthenticated: true,
      user: { id: '1', username: 'alice', role: 'user' },
    });
    const { result, unmount } = renderHook(() => useWebSocket(), { wrapper });
    await waitFor(() => expect(MockWebSocket.instances.length).toBe(1));
    const ws = MockWebSocket.latest();
    act(() => {
      ws.simulateOpen();
      ws.onmessage?.({ data: 'not-json' });
      ws.simulateError();
    });
    act(() => {
      result.current.reconnect();
      result.current.disconnect();
    });
    unmount();
  });

  it('uses wss when page is https', async () => {
    vi.spyOn(window.location, 'protocol', 'get').mockReturnValue('https:');
    renderHook(() => useWebSocket(), { wrapper });
    await waitFor(() => expect(MockWebSocket.latest().url.startsWith('wss:')).toBe(true));
  });

  it('throws outside provider', () => {
    expect(() => renderHook(() => useWebSocket())).toThrow(
      'useWebSocket must be used within WebSocketProvider',
    );
  });
});
