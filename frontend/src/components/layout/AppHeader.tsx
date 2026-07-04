import { Link } from 'react-router-dom';
import type { NavLink } from '../../constants/navigation';
import { useAuth } from '../../hooks/useAuth';

type Props = {
  links?: NavLink[];
  onLogout?: () => void;
};

export default function AppHeader({ links = [], onLogout }: Props) {
  const { user } = useAuth();

  return (
    <header className="app-header">
      <div className="brand">NumDuel</div>
      <nav className="nav-links">
        {links.map((link) => (
          <Link key={link.to} to={link.to}>
            {link.label}
          </Link>
        ))}
      </nav>
      <div className="header-actions">
        {user && <span className="username">{user.username}</span>}
        {onLogout && (
          <button type="button" className="btn-secondary" onClick={onLogout}>
            ログアウト
          </button>
        )}
      </div>
    </header>
  );
}
