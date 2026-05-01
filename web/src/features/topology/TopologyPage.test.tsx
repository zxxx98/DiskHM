import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';

import { TopologyPage } from './TopologyPage';

describe('TopologyPage', () => {
  afterEach(() => {
    cleanup();
  });

  it('renders topology nodes and edge count', () => {
    render(
      <TopologyPage
        edges={[{ from: 'host-1', to: 'disk-sda' }]}
        nodes={[{ id: 'disk-sda', label: '/dev/sda' }]}
      />,
    );

    expect(screen.getByRole('heading', { name: 'Topology' })).toBeInTheDocument();
    expect(screen.getByText('/dev/sda')).toBeInTheDocument();
    expect(screen.getByText('1 edge')).toBeInTheDocument();
  });

  it('renders an empty-state message when topology data is not available yet', () => {
    render(<TopologyPage edges={[]} nodes={[]} />);

    expect(screen.getByRole('heading', { name: 'Topology' })).toBeInTheDocument();
    expect(screen.getByText('0 edges')).toBeInTheDocument();
    expect(screen.getByText('No topology nodes reported yet.')).toBeInTheDocument();
  });
});
