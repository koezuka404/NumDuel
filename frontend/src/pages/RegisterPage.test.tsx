import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
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

function renderPage() {
  return render(
    <MemoryRouter>
      <ToastProvider>
        <RegisterPage />
      </ToastProvider>
    </MemoryRouter>,
  );
}

function fillValidForm() {
  fireEvent.change(screen.getByLabelText('ユーザー名'), { target: { value: 'alice' } });
  fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'a@test.local' } });
  fireEvent.change(screen.getAllByLabelText('パスワード')[0], { target: { value: 'password123' } });
  fireEvent.change(screen.getByLabelText('パスワード確認'), { target: { value: 'password123' } });
}

describe('RegisterPage', () => {
  afterEach(() => {
    cleanup();
    navigate.mockReset();
    apiFetchMock.mockReset();
  });

  it('shows validation error for short username', async () => {
    renderPage();
    fillValidForm();
    fireEvent.change(screen.getByLabelText('ユーザー名'), { target: { value: 'ab' } });
    fireEvent.click(screen.getByRole('button', { name: '登録' }));
    await waitFor(() => expect(screen.getByText(/3〜50文字/)).toBeInTheDocument());
  });

  it('registers successfully', async () => {
    apiFetchMock.mockResolvedValueOnce(undefined);
    renderPage();
    fillValidForm();
    fireEvent.click(screen.getByRole('button', { name: '登録' }));
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/login'));
  });

  it('shows api error', async () => {
    apiFetchMock.mockRejectedValueOnce(new ApiError('validation_error', 'duplicate', 400));
    renderPage();
    fillValidForm();
    fireEvent.click(screen.getByRole('button', { name: '登録' }));
    await waitFor(() => expect(screen.getByText('duplicate')).toBeInTheDocument());
  });
});
