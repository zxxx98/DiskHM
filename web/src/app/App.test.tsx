import { cleanup, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { App } from './App';

function renderApp() {
  const queryClient = new QueryClient();

  return render(
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>,
  );
}

describe('App', () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    window.history.pushState({}, '', '/');
  });

  it('renders the DiskHM login shell', () => {
    renderApp();

    expect(screen.getByRole('heading', { name: 'DiskHM' })).toBeInTheDocument();
    expect(screen.getByLabelText('Token')).toBeInTheDocument();
  });

  it('renders the topology route when the location changes', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response('{"nodes":[],"edges":[]}', {
          status: 200,
          headers: {
            'Content-Type': 'application/json',
          },
        }),
      ),
    );
    window.history.pushState({}, '', '/topology');

    renderApp();

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Topology' })).toBeInTheDocument();
    });
    expect(screen.queryByRole('heading', { name: 'DiskHM' })).not.toBeInTheDocument();
    expect(screen.queryByText('View scaffold pending backend integration.')).not.toBeInTheDocument();
    expect(screen.getByText('No topology nodes reported yet.')).toBeInTheDocument();
  });

  it('renders real disks from the API on the disks route', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          '{"items":[{"id":"disk-sda","name":"sda","path":"/dev/sda","model":"WD Red","powerState":"unknown","refreshFreshness":"cached","unsupported":false,"mounts":["/srv/data"]}]}',
          {
            status: 200,
            headers: {
              'Content-Type': 'application/json',
            },
          },
        ),
      ),
    );
    window.history.pushState({}, '', '/disks');

    renderApp();

    expect(await screen.findByText('WD Red')).toBeInTheDocument();
    expect(screen.getByText('/srv/data')).toBeInTheDocument();
  });

  it('shows a disk loading error when the disks API request fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response('{"code":"internal_error","message":"failed to list disks"}', {
          status: 500,
          headers: {
            'Content-Type': 'application/json',
          },
        }),
      ),
    );
    window.history.pushState({}, '', '/disks');

    renderApp();

    expect(await screen.findByRole('alert')).toHaveTextContent('HTTP 500');
  });
});
