import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import LoginPage from './LoginPage';
import { ApiError } from '../api/client';

const login = vi.fn();

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({
    user: null,
    isAuthenticated: false,
    isLoading: false,
    login,
    logout: vi.fn(),
    refreshUser: vi.fn(),
  }),
}));

describe('LoginPage', () => {
  it('shows unauthorized message', async () => {
    login.mockRejectedValueOnce(new ApiError('unauthorized', 'bad credentials', 401));

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'user@test.local' } });
    fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));

    await waitFor(() => {
      expect(screen.getByText('メールまたはパスワードが正しくありません')).toBeInTheDocument();
    });
  });
});
