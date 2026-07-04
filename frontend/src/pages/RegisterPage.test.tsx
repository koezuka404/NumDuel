import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { ApiError, apiFetch } from '../api/client';
import { ToastProvider } from '../hooks/useToast';
import RegisterPage from './RegisterPage';

const navigate = vi.fn();
const apiFetchMock = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => navigate };
});

vi.mock('../api/client', async () => {
  const actual = await vi.importActual<typeof import('../api/client')>('../api/client');
  return { ...actual, apiFetch: (...args: Parameters<typeof apiFetch>) => apiFetchMock(...args) };
});

function renderRegisterPage() {
  return render(
    <MemoryRouter>
      <ToastProvider>
        <RegisterPage />
      </ToastProvider>
    </MemoryRouter>,
  );
}

describe('RegisterPage', () => {
  it('shows validation error for short username', async () => {
    renderRegisterPage();
    fireEvent.change(screen.getByLabelText('ユーザー名'), { target: { value: 'ab' } });
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'a@test.local' } });
    fireEvent.change(screen.getAllByLabelText('パスワード')[0], { target: { value: 'password123' } });
    fireEvent.change(screen.getByLabelText('パスワード確認'), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: '登録' }));
    await waitFor(() => {
      expect(screen.getByText(/3〜50文字/)).toBeInTheDocument();
    });
  });

  it('shows password mismatch error', async () => {
    renderRegisterPage();
    fireEvent.change(screen.getByLabelText('ユーザー名'), { target: { value: 'alice' } });
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'a@test.local' } });
    fireEvent.change(screen.getAllByLabelText('パスワード')[0], { target: { value: 'password123' } });
    fireEvent.change(screen.getByLabelText('パスワード確認'), { target: { value: 'password124' } });
    fireEvent.click(screen.getByRole('button', { name: '登録' }));
    await waitFor(() => {
      expect(screen.getByText('パスワード確認が一致しません')).toBeInTheDocument();
    });
  });

  it('registers successfully', async () => {
    apiFetchMock.mockResolvedValueOnce(undefined);
    renderRegisterPage();
    fireEvent.change(screen.getByLabelText('ユーザー名'), { target: { value: 'alice' } });
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'a@test.local' } });
    fireEvent.change(screen.getAllByLabelText('パスワード')[0], { target: { value: 'password123' } });
    fireEvent.change(screen.getByLabelText('パスワード確認'), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: '登録' }));
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/login'));
    expect(screen.getByText('登録完了')).toBeInTheDocument();
  });

  it('handles api errors', async () => {
    apiFetchMock.mockRejectedValueOnce(new ApiError('validation_error', 'duplicate', 400));
    renderRegisterPage();
    fireEvent.change(screen.getByLabelText('ユーザー名'), { target: { value: 'alice' } });
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'a@test.local' } });
    fireEvent.change(screen.getAllByLabelText('パスワード')[0], { target: { value: 'password123' } });
    fireEvent.change(screen.getByLabelText('パスワード確認'), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: '登録' }));
    await waitFor(() => expect(screen.getByText('duplicate')).toBeInTheDocument());

    apiFetchMock.mockRejectedValueOnce(new ApiError('rate_limit_exceeded', 'too many', 429));
    fireEvent.click(screen.getByRole('button', { name: '登録' }));
    await waitFor(() => expect(screen.getByText('too many')).toBeInTheDocument());

    apiFetchMock.mockRejectedValueOnce(new Error('network'));
    fireEvent.click(screen.getByRole('button', { name: '登録' }));
    await waitFor(() => expect(screen.getByText('登録に失敗しました')).toBeInTheDocument();
  });
});
