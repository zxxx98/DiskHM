import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { DiskActions } from './DiskActions';
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

  it('disables the sleep actions for unsupported disks', () => {
    render(
      <DiskTablePage
        disks={[
          {
            id: 'disk-sdb',
            name: 'sdb',
            model: 'Seagate IronWolf',
            powerState: 'active',
            refreshFreshness: 'cached',
            unsupported: true,
          },
        ]}
      />,
    );

    expect(screen.getByRole('button', { name: 'Sleep now' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Sleep in 30m' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Refresh (wake disk)' })).toBeInTheDocument();
  });
});

describe('DiskActions', () => {
  afterEach(() => {
    cleanup();
  });

  it('uses the explicit action callbacks', () => {
    const onSleepNow = vi.fn();
    const onSleepLater = vi.fn();
    const onWakeRefresh = vi.fn();

    render(
      <DiskActions
        disableSleepNow={false}
        disableSleepLater={false}
        disableWakeRefresh={false}
        onSleepNow={onSleepNow}
        onSleepLater={onSleepLater}
        onWakeRefresh={onWakeRefresh}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Sleep now' }));
    fireEvent.click(screen.getByRole('button', { name: 'Sleep in 30m' }));
    fireEvent.click(screen.getByRole('button', { name: 'Refresh (wake disk)' }));

    expect(onSleepNow).toHaveBeenCalledTimes(1);
    expect(onSleepLater).toHaveBeenCalledTimes(1);
    expect(onWakeRefresh).toHaveBeenCalledTimes(1);
  });
});
