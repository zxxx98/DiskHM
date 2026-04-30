import { useCallback, useState } from 'react';

import { useEventStream } from './useEventStream';

type EventItem = {
  id: string;
  message: string;
};

type EventsPageProps = {
  events?: EventItem[];
};

export function EventsPage({ events = [] }: EventsPageProps) {
  const [streamEvents, setStreamEvents] = useState(events);

  const handleMessage = useCallback((event: MessageEvent<string>) => {
    setStreamEvents((current) => [
      ...current,
      {
        id: `${current.length}-${event.data}`,
        message: event.data,
      },
    ]);
  }, []);

  useEventStream(handleMessage);

  return (
    <main className="disk-page">
      <section className="disk-page__panel" aria-labelledby="events-page-title">
        <div className="disk-page__header">
          <div>
            <p className="disk-page__eyebrow">Activity feed</p>
            <h1 id="events-page-title">Events</h1>
          </div>
          <p className="disk-page__copy">Recent operational messages stream here as they arrive.</p>
        </div>

        <ul aria-label="Event messages">
          {streamEvents.map((event) => (
            <li key={event.id}>{event.message}</li>
          ))}
        </ul>
      </section>
    </main>
  );
}
