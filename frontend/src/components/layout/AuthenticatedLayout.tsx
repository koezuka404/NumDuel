import type { ReactNode } from 'react';
import type { NavLink } from '../../constants/navigation';
import AppHeader from './AppHeader';
import AppShell from './AppShell';
import { useLogout } from '../../hooks/useLogout';

type Props = {
  children: ReactNode;
  links?: NavLink[];
};

export default function AuthenticatedLayout({ children, links = [] }: Props) {
  const handleLogout = useLogout();

  return (
    <AppShell header={<AppHeader links={links} onLogout={handleLogout} />}>
      {children}
    </AppShell>
  );
}
