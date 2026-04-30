import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { EventsPage } from './EventsPage';
import { useEventStream } from './useEventStream';

describe('EventsPage', () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it('renders event messages', () => {
    vi.stubGlobal(
      'EventSource',
      vi.fn(() => ({
        close: vi.fn(),
        onmessage: null,
      })),
    );

    render(<EventsPage events={[{ id: 'evt-1', message: 'disk queued' }]} />);

    expect(screen.getByRole('heading', { name: 'Events' })).toBeInTheDocument();
    expect(screen.getByText('disk queued')).toBeInTheDocument();
  });
});

describe('useEventStream', () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it('opens the event stream, forwards messages, and closes on cleanup', () => {
    const close = vi.fn();
    let source: { onmessage: ((event: MessageEvent<string>) => void) | null } | null = null;

    const EventSourceMock = vi.fn(
      (
        url: string,
        options: {
          withCredentials: boolean;
        },
      ) => {
        source = { onmessage: null };

        return {
          close,
          get onmessage() {
            return source?.onmessage ?? null;
          },
          set onmessage(handler: ((event: MessageEvent<string>) => void) | null) {
            if (source) {
              source.onmessage = handler;
            }
          },
          url,
          withCredentials: options.withCredentials,
        };
      },
    );

    vi.stubGlobal('EventSource', EventSourceMock);

    const received: string[] = [];

    function Harness() {
      useEventStream((event) => {
        received.push(event.data);
      });

      return null;
    }

    const view = render(<Harness />);

    expect(EventSourceMock).toHaveBeenCalledWith('/api/events/stream', { withCredentials: true });

    source?.onmessage?.({ data: 'disk queued' } as MessageEvent<string>);
    expect(received).toEqual(['disk queued']);

    view.unmount();

    expect(close).toHaveBeenCalledTimes(1);
  });
});
