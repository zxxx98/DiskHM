type EventItem = {
  id: string;
  message: string;
};

type EventsPageProps = {
  events?: EventItem[];
};

export function EventsPage({ events = [] }: EventsPageProps) {
  return (
    <main className="disk-page">
      <section className="disk-page__panel" aria-labelledby="events-page-title">
        <div className="disk-page__header">
          <div>
            <p className="disk-page__eyebrow">Activity feed</p>
            <h1 id="events-page-title">Events</h1>
          </div>
          <p className="disk-page__copy">Live stream hookup stays off until event transport and retention are wired.</p>
        </div>

        <ul aria-label="Event messages">
          {events.length === 0 ? <li>No events received yet.</li> : null}
          {events.map((event) => (
            <li key={event.id}>{event.message}</li>
          ))}
        </ul>
      </section>
    </main>
  );
}
