import type { RouteObject } from 'react-router-dom';

import { LoginPage } from '../features/session/LoginPage';

function ShellPlaceholder({ title }: { title: string }) {
  return (
    <section className="shell-placeholder" aria-label={title}>
      <h1>{title}</h1>
      <p>View scaffold pending backend integration.</p>
    </section>
  );
}

export const routes: RouteObject[] = [
  {
    path: '/',
    element: <LoginPage />,
  },
  {
    path: '/topology',
    element: <ShellPlaceholder title="Topology" />,
  },
  {
    path: '/events',
    element: <ShellPlaceholder title="Events" />,
  },
  {
    path: '/settings',
    element: <ShellPlaceholder title="Settings" />,
  },
];
