import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import { notifyUnauthorized } from '../api/client';
import { resolveApiBaseURL, resolveWsBaseURL } from '../lib/apiBase';
import type { WSMessage } from '../types/dto';
import { useAuth } from './useAuth';

type WSListener = (msg: WSMessage) => void;

type WebSocketContextValue = {
  connected: boolean;
  connecting: boolean;
  send: (msg: Record<string, unknown>) => void;
  subscribe: (listener: WSListener) => () => void;
  disconnect: () => void;
  reconnect: () => void;
};

const WebSocketContext = createContext<WebSocketContextValue | null>(null);

const API_BASE_URL = resolveApiBaseURL();
const WS_BASE_URL = resolveWsBaseURL();

async function fetchWsTicket(): Promise<string> {
  try {
    const res = await fetch(`${API_BASE_URL}/auth/ws-ticket`, { credentials: 'include' });
    if (!res.ok) {
      return '';
    }
    const json = await res.json();
    return json?.data?.ticket ?? json?.ticket ?? '';
  } catch {
    return '';
  }
}

export function WebSocketProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated, user } = useAuth();
  const [connected, setConnected] = useState(false);
  const [connecting, setConnecting] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const listenersRef = useRef(new Set<WSListener>());
  const pingRef = useRef<number | null>(null);
  const reconnectAttemptRef = useRef(0);
  const reconnectTimerRef = useRef<number | null>(null);
  const shouldConnectRef = useRef(false);

  const emit = useCallback((msg: WSMessage) => {
    listenersRef.current.forEach((listener) => listener(msg));
  }, []);

  const clearPing = useCallback(() => {
    if (pingRef.current !== null) {
      window.clearInterval(pingRef.current);
      pingRef.current = null;
    }
  }, []);

  const disconnect = useCallback(() => {
    shouldConnectRef.current = false;
    clearPing();
    if (reconnectTimerRef.current !== null) {
      window.clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
    reconnectAttemptRef.current = 0;
    wsRef.current?.close();
    wsRef.current = null;
    setConnected(false);
    setConnecting(false);
  }, [clearPing]);

  const handleWsError = useCallback(
    async (code: string) => {
      if (code === 'token_expired') {
        await fetch(`${API_BASE_URL}/auth/refresh`, { method: 'POST', credentials: 'include' });
        reconnectAttemptRef.current = 0;
        wsRef.current?.close();
        return;
      }
      if (code === 'unauthorized') {
        disconnect();
        notifyUnauthorized();
      }
    },
    [disconnect],
  );

  const connect = useCallback(async () => {
    if (!shouldConnectRef.current || wsRef.current) {
      return;
    }
    setConnecting(true);

    const ticket = await fetchWsTicket();

    if (!shouldConnectRef.current) {
      setConnecting(false);
      return;
    }

    const ws = new WebSocket(WS_BASE_URL);
    wsRef.current = ws;

    ws.onopen = () => {
      reconnectAttemptRef.current = 0;
      ws.send(JSON.stringify({ type: 'AUTH', ticket }));
      clearPing();
      pingRef.current = window.setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: 'PING' }));
        }
      }, 30_000);
    };

    ws.onmessage = (event) => {
      let msg: WSMessage;
      try {
        msg = JSON.parse(event.data) as WSMessage;
      } catch {
        return;
      }
      if (msg.type === 'AUTH_OK') {
        setConnected(true);
        setConnecting(false);
      }
      if (msg.type === 'ERROR') {
        const code = String(msg.data?.code ?? '');
        void handleWsError(code);
      }
      emit(msg);
    };

    ws.onclose = () => {
      setConnected(false);
      setConnecting(false);
      clearPing();
      wsRef.current = null;
      if (!shouldConnectRef.current) {
        return;
      }
      if (reconnectAttemptRef.current >= 5) {
        emit({ type: 'RECONNECT_FAILED' });
        return;
      }
      const delay = 1000 * 2 ** reconnectAttemptRef.current;
      reconnectAttemptRef.current += 1;
      reconnectTimerRef.current = window.setTimeout(() => connect(), delay);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [clearPing, emit, handleWsError]);

  const reconnect = useCallback(() => {
    wsRef.current?.close();
    wsRef.current = null;
    void connect();
  }, [connect]);

  useEffect(() => {
    const enabled = isAuthenticated && user?.role === 'user';
    if (enabled) {
      shouldConnectRef.current = true;
      void connect();
    } else {
      disconnect();
    }
    return () => disconnect();
  }, [isAuthenticated, user?.role, connect, disconnect]);

  const send = useCallback((msg: Record<string, unknown>) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg));
    }
  }, []);

  const subscribe = useCallback((listener: WSListener) => {
    listenersRef.current.add(listener);
    return () => listenersRef.current.delete(listener);
  }, []);

  const value = useMemo(
    () => ({ connected, connecting, send, subscribe, disconnect, reconnect }),
    [connected, connecting, send, subscribe, disconnect, reconnect],
  );

  return <WebSocketContext.Provider value={value}>{children}</WebSocketContext.Provider>;
}

export function useWebSocket(): WebSocketContextValue {
  const ctx = useContext(WebSocketContext);
  if (!ctx) {
    throw new Error('useWebSocket must be used within WebSocketProvider');
  }
  return ctx;
}