import { BrowserRouter, useRoutes } from 'react-router-dom';

import { routes } from './routes';
import '../styles/app.css';

function AppRoutes() {
  return useRoutes(routes);
}

export function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  );
}
