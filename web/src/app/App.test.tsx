import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';

import { App } from './App';

describe('App', () => {
  afterEach(() => {
    cleanup();
    window.history.pushState({}, '', '/');
  });

  it('renders the DiskHM login shell', () => {
    render(<App />);

    expect(screen.getByRole('heading', { name: 'DiskHM' })).toBeInTheDocument();
    expect(screen.getByLabelText('Token')).toBeInTheDocument();
  });

  it('renders the login shell directly regardless of location', () => {
    window.history.pushState({}, '', '/topology');

    render(<App />);

    expect(screen.getByRole('heading', { name: 'DiskHM' })).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Topology' })).not.toBeInTheDocument();
  });
});
