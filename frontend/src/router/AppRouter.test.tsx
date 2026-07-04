import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import AppRouter from '../router/AppRouter';

const useAuth = vi.fn();

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => useAuth(),
}));

vi.mock('../pages/MatchingPage', () => ({ default: () => <div>Matching</div> }));
vi.mock('../pages/LoginPage', () => ({ default: () => <h1>ログイン</h1> }));

function auth(overrides: Record<string, unknown> = {}) {
  return {
    user: null,
    isAuthenticated: false,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
    refreshUser: vi.fn(),
    ...overrides,
  };
}

describe('AppRouter', () => {
  afterEach(() => vi.clearAllMocks());

  it('redirects unauthenticated users to login', () => {
    useAuth.mockReturnValue(auth());
    render(
      <MemoryRouter initialEntries={['/matching']}>
        <AppRouter />
      </MemoryRouter>,
    );
    expect(screen.getByRole('heading', { name: 'ログイン' })).toBeInTheDocument();
  });

  it('renders matching for authenticated users', () => {
    useAuth.mockReturnValue(
      auth({ isAuthenticated: true, user: { id: '1', username: 'alice', role: 'user' } }),
    );
    render(
      <MemoryRouter initialEntries={['/matching']}>
        <AppRouter />
      </MemoryRouter>,
    );
    expect(screen.getByText('Matching')).toBeInTheDocument();
  });
});
