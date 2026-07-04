import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import AppRouter from '../router/AppRouter';

const useAuth = vi.fn();

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => useAuth(),
}));

vi.mock('../pages/MatchingPage', () => ({ default: () => <div>Matching</div> }));
vi.mock('../pages/GamePage', () => ({ default: () => <div>Game</div> }));
vi.mock('../pages/RankingPage', () => ({ default: () => <div>Ranking</div> }));
vi.mock('../pages/ProfilePage', () => ({ default: () => <div>Profile</div> }));
vi.mock('../pages/AdminPage', () => ({ default: () => <div>Admin</div> }));
vi.mock('../pages/LoginPage', () => ({ default: () => <h1>ログイン</h1> }));
vi.mock('../pages/RegisterPage', () => ({ default: () => <h1>新規登録</h1> }));

function authState(overrides: Record<string, unknown>) {
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
  it('redirects unauthenticated users to login for protected routes', () => {
    useAuth.mockReturnValue(authState({}));
    render(
      <MemoryRouter initialEntries={['/matching']}>
        <AppRouter />
      </MemoryRouter>,
    );
    expect(screen.getByRole('heading', { name: 'ログイン' })).toBeInTheDocument();
  });

  it('renders matching for authenticated users', () => {
    useAuth.mockReturnValue(
      authState({
        isAuthenticated: true,
        user: { id: '1', username: 'alice', role: 'user' },
      }),
    );
    render(
      <MemoryRouter initialEntries={['/matching']}>
        <AppRouter />
      </MemoryRouter>,
    );
    expect(screen.getByText('Matching')).toBeInTheDocument();
  });

  it('redirects master users away from user routes and guests away from auth pages', () => {
    useAuth.mockReturnValue(
      authState({
        isAuthenticated: true,
        user: { id: '1', username: 'admin', role: 'master' },
      }),
    );
    render(
      <MemoryRouter initialEntries={['/matching']}>
        <AppRouter />
      </MemoryRouter>,
    );
    expect(screen.queryByText('Matching')).not.toBeInTheDocument();

    useAuth.mockReturnValue(
      authState({
        isAuthenticated: true,
        user: { id: '1', username: 'admin', role: 'master' },
      }),
    );
    render(
      <MemoryRouter initialEntries={['/admin']}>
        <AppRouter />
      </MemoryRouter>,
    );
    expect(screen.getByText('Admin')).toBeInTheDocument();

    useAuth.mockReturnValue(
      authState({
        isAuthenticated: true,
        user: { id: '1', username: 'alice', role: 'user' },
      }),
    );
    render(
      <MemoryRouter initialEntries={['/login']}>
        <AppRouter />
      </MemoryRouter>,
    );
    expect(screen.queryByRole('heading', { name: 'ログイン' })).not.toBeInTheDocument();
  });

  it('renders other protected pages and catch-all redirect', () => {
    useAuth.mockReturnValue(
      authState({
        isAuthenticated: true,
        user: { id: '1', username: 'alice', role: 'user' },
      }),
    );
    render(
      <MemoryRouter initialEntries={['/game/1']}>
        <AppRouter />
      </MemoryRouter>,
    );
    expect(screen.getByText('Game')).toBeInTheDocument();

    render(
      <MemoryRouter initialEntries={['/ranking']}>
        <AppRouter />
      </MemoryRouter>,
    );
    expect(screen.getByText('Ranking')).toBeInTheDocument();

    render(
      <MemoryRouter initialEntries={['/profile']}>
        <AppRouter />
      </MemoryRouter>,
    );
    expect(screen.getByText('Profile')).toBeInTheDocument();

    useAuth.mockReturnValue(authState({}));
    render(
      <MemoryRouter initialEntries={['/unknown']}>
        <AppRouter />
      </MemoryRouter>,
    );
    expect(screen.getByRole('heading', { name: 'ログイン' })).toBeInTheDocument();
  });
});
