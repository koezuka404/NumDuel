import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { GuestOnly, RequireAuth, RequireMaster, RequireUser } from './guards';

const useAuth = vi.fn();

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => useAuth(),
}));

function renderGuard(element: JSX.Element, path = '/') {
  return render(
    <MemoryRouter initialEntries={[path]}>
      {element}
    </MemoryRouter>,
  );
}

afterEach(() => {
  cleanup();
});

describe('RequireAuth', () => {
  it('shows loading screen while auth is loading', () => {
    useAuth.mockReturnValue({ isAuthenticated: false, isLoading: true, user: null });
    renderGuard(
      <RequireAuth>
        <div>protected</div>
      </RequireAuth>,
    );
    expect(screen.getAllByLabelText('読み込み中').length).toBeGreaterThan(0);
  });

  it('redirects unauthenticated users to login', () => {
    useAuth.mockReturnValue({ isAuthenticated: false, isLoading: false, user: null });
    renderGuard(
      <RequireAuth>
        <div>protected</div>
      </RequireAuth>,
    );
    expect(screen.queryByText('protected')).not.toBeInTheDocument();
  });

  it('renders children when authenticated', () => {
    useAuth.mockReturnValue({
      isAuthenticated: true,
      isLoading: false,
      user: { id: '1', username: 'alice', role: 'user' },
    });
    renderGuard(
      <RequireAuth>
        <div>protected</div>
      </RequireAuth>,
    );
    expect(screen.getByText('protected')).toBeInTheDocument();
  });
});

describe('RequireUser', () => {
  it('shows loading screen while auth is loading', () => {
    useAuth.mockReturnValue({ isLoading: true, user: null });
    renderGuard(
      <RequireUser>
        <div>user-only</div>
      </RequireUser>,
    );
    expect(screen.getAllByLabelText('読み込み中').length).toBeGreaterThan(0);
  });

  it('redirects master users to admin', () => {
    useAuth.mockReturnValue({
      isLoading: false,
      user: { id: '1', username: 'admin', role: 'master' },
    });
    renderGuard(
      <RequireUser>
        <div>user-only</div>
      </RequireUser>,
    );
    expect(screen.queryByText('user-only')).not.toBeInTheDocument();
  });

  it('renders children for regular users', () => {
    useAuth.mockReturnValue({
      isLoading: false,
      user: { id: '1', username: 'alice', role: 'user' },
    });
    renderGuard(
      <RequireUser>
        <div>user-only</div>
      </RequireUser>,
    );
    expect(screen.getByText('user-only')).toBeInTheDocument();
  });
});

describe('RequireMaster', () => {
  it('shows loading screen while auth is loading', () => {
    useAuth.mockReturnValue({ isLoading: true, user: null });
    renderGuard(
      <RequireMaster>
        <div>admin-only</div>
      </RequireMaster>,
    );
    expect(screen.getAllByLabelText('読み込み中').length).toBeGreaterThan(0);
  });

  it('redirects non-master users to matching', () => {
    useAuth.mockReturnValue({
      isLoading: false,
      user: { id: '1', username: 'alice', role: 'user' },
    });
    renderGuard(
      <RequireMaster>
        <div>admin-only</div>
      </RequireMaster>,
    );
    expect(screen.queryByText('admin-only')).not.toBeInTheDocument();
  });

  it('renders children for master users', () => {
    useAuth.mockReturnValue({
      isLoading: false,
      user: { id: '1', username: 'admin', role: 'master' },
    });
    renderGuard(
      <RequireMaster>
        <div>admin-only</div>
      </RequireMaster>,
    );
    expect(screen.getByText('admin-only')).toBeInTheDocument();
  });
});

describe('GuestOnly', () => {
  it('shows loading screen while auth is loading', () => {
    useAuth.mockReturnValue({ isAuthenticated: false, isLoading: true, user: null });
    renderGuard(
      <GuestOnly>
        <div>guest</div>
      </GuestOnly>,
    );
    expect(screen.getAllByLabelText('読み込み中').length).toBeGreaterThan(0);
  });

  it('redirects authenticated users to home path', () => {
    useAuth.mockReturnValue({
      isAuthenticated: true,
      isLoading: false,
      user: { id: '1', username: 'alice', role: 'user' },
    });
    renderGuard(
      <GuestOnly>
        <div>guest</div>
      </GuestOnly>,
    );
    expect(screen.queryByText('guest')).not.toBeInTheDocument();
  });

  it('renders children for guests', () => {
    useAuth.mockReturnValue({
      isAuthenticated: false,
      isLoading: false,
      user: null,
    });
    renderGuard(
      <GuestOnly>
        <div>guest</div>
      </GuestOnly>,
    );
    expect(screen.getByText('guest')).toBeInTheDocument();
  });
});
