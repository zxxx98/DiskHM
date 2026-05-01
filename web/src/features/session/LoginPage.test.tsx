import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { getCsrfToken, setCsrfToken } from '../../lib/http';
import { LoginPage } from './LoginPage';

describe('LoginPage', () => {
  afterEach(() => {
    cleanup();
    setCsrfToken(null);
    vi.restoreAllMocks();
  });

  it('submits the token to the login endpoint and navigates to disks on success', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(null, {
        status: 204,
        headers: {
          'X-CSRF-Token': 'csrf-token',
        },
      }),
    );
    vi.stubGlobal('fetch', fetchMock);

    render(
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route path="/" element={<LoginPage />} />
          <Route path="/disks" element={<h1>Disks</h1>} />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText('Token'), {
      target: { value: 'dev-token' },
    });
    fireEvent.submit(screen.getByRole('button', { name: 'Sign in' }).closest('form') as HTMLFormElement);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(1);
    });

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/session/login',
      expect.objectContaining({
        credentials: 'include',
        method: 'POST',
      }),
    );
    expect(await screen.findByRole('heading', { name: 'Disks' })).toBeInTheDocument();
    expect(getCsrfToken()).toBe('csrf-token');
  });
});
