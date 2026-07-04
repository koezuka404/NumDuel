import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import App from './App';

vi.mock('./router/AppRouter', () => ({
  default: () => <div>router</div>,
}));

vi.mock('./hooks/useAuth', () => ({
  AuthProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('./hooks/useWebSocket', () => ({
  WebSocketProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('./hooks/useToast', () => ({
  ToastProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

describe('App', () => {
  it('renders router inside providers', () => {
    render(<App />);
    expect(screen.getByText('router')).toBeInTheDocument();
  });
});
