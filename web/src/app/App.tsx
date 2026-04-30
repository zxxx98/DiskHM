import { useMemo } from 'react';
import { QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider, createBrowserRouter } from 'react-router-dom';

import { queryClient } from '../lib/query';
import { routes } from './routes';
import { LoginPage } from '../features/session/LoginPage';
import '../styles/app.css';

export function App() {
  const router = useMemo(() => createBrowserRouter(routes), []);

  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} fallbackElement={<LoginPage />} />
    </QueryClientProvider>
  );
}
