import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';

import { App } from './App';

describe('App', () => {
  afterEach(() => {
    window.history.pushState({}, '', '/');
  });

  it('renders the DiskHM login shell', () => {
    render(<App />);

    expect(screen.getByRole('heading', { name: 'DiskHM' })).toBeInTheDocument();
    expect(screen.getByLabelText('Token')).toBeInTheDocument();
  });

  it('renders the topology route through the shared router scaffold', () => {
    window.history.pushState({}, '', '/topology');

    render(<App />);

    expect(screen.getByRole('heading', { name: 'Topology' })).toBeInTheDocument();
  });
});
