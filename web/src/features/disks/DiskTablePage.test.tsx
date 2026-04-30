import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';

import { DiskTablePage } from './DiskTablePage';

describe('DiskTablePage', () => {
  afterEach(() => {
    cleanup();
  });

  it('renders the disk inventory with wake-capable refresh action', () => {
    render(
      <DiskTablePage
        disks={[
          {
            id: 'disk-sda',
            name: 'sda',
            model: 'WD Red',
            powerState: 'sleeping',
            refreshFreshness: 'cached',
            unsupported: false,
          },
        ]}
      />,
    );

    expect(screen.getByText('WD Red')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Refresh (wake disk)' })).toBeInTheDocument();
  });
});
