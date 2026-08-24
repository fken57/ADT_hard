import { Link, NavLink } from 'react-router-dom';
import { ReactNode } from 'react';

export function Layout({ children }: { children: ReactNode }) {
  return (
    <div className="app-shell">
      <header className="site-header">
        <Link to="/" className="brand">AtCoder <strong>Shojin</strong></Link>
        <nav aria-label="主要ナビゲーション"><NavLink to="/history">履歴</NavLink></nav>
      </header>
      <main className="page-container">{children}</main>
    </div>
  );
}
