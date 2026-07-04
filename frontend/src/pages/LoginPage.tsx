import { FormEvent, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ApiError } from '../api/client';
import AppShell from '../components/layout/AppShell';
import FormField, { FormError } from '../components/ui/FormField';
import { useAuth } from '../hooks/useAuth';
import { homePathForRole } from '../lib/routes';
import { validateLoginEmail, validatePassword } from '../lib/validation';

export default function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [formError, setFormError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault();
    const emailError = validateLoginEmail(email);
    const passwordError = validatePassword(password);
    if (emailError || passwordError) {
      setFormError(emailError ?? passwordError ?? '');
      return;
    }

    setSubmitting(true);
    setFormError('');
    try {
      const user = await login(email, password);
      navigate(homePathForRole(user.role));
    } catch (err) {
      if (err instanceof ApiError && err.code === 'unauthorized') {
        setFormError('メールまたはパスワードが正しくありません');
      } else if (err instanceof ApiError) {
        setFormError(err.message);
      } else {
        setFormError('ログインに失敗しました');
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <AppShell narrow>
      <h1>ログイン</h1>
      {formError && <FormError message={formError} />}
      <form onSubmit={onSubmit} className="form-card">
        <FormField label="メールアドレス" type="email" value={email} onChange={setEmail} required />
        <FormField label="パスワード" type="password" value={password} onChange={setPassword} required />
        <button type="submit" className="btn-primary btn-block" disabled={submitting}>
          {submitting ? '処理中…' : 'ログイン'}
        </button>
      </form>
      <p className="auth-footer">
        <Link to="/register">新規登録はこちら</Link>
      </p>
    </AppShell>
  );
}
