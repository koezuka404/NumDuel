import type { ReactNode } from 'react';

type Props = {
  children: ReactNode;
  header?: ReactNode;
  narrow?: boolean;
};

export default function AppShell({ children, header, narrow = false }: Props) {
  return (
    <div className="app-shell">
      {header ?? (
        <header className="app-shell__header app-shell__header--brand-only">
          <span className="app-shell__brand">NumDuel</span>
        </header>
      )}
      <main className={narrow ? 'app-shell__main app-shell__main--narrow' : 'app-shell__main'}>{children}</main>
    </div>
  );
}
