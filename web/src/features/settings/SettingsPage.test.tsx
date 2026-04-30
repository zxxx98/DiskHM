import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';

import { SettingsPage } from './SettingsPage';

describe('SettingsPage', () => {
  afterEach(() => {
    cleanup();
  });

  it('renders the quiet grace seconds input', () => {
    render(<SettingsPage quietGraceSeconds={10} />);

    expect(screen.getByRole('heading', { name: 'Settings' })).toBeInTheDocument();
    expect(screen.getByLabelText('Quiet grace seconds')).toBeDisabled();
    expect(screen.getByText('Settings stay read-only until the save flow and validation rules are in place.')).toBeInTheDocument();
  });
});
