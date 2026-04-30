type SettingsPageProps = {
  quietGraceSeconds: number;
};

export function SettingsPage({ quietGraceSeconds }: SettingsPageProps) {
  return (
    <main className="disk-page">
      <section className="disk-page__panel" aria-labelledby="settings-page-title">
        <div className="disk-page__header">
          <div>
            <p className="disk-page__eyebrow">Policy</p>
            <h1 id="settings-page-title">Settings</h1>
          </div>
          <p className="disk-page__copy">Adjust quiet-period timing without touching the auth or disk control flow.</p>
        </div>

        <form>
          <div className="field">
            <label className="field__label" htmlFor="quiet-grace-seconds">
              Quiet grace seconds
            </label>
            <input
              className="field__input"
              defaultValue={quietGraceSeconds}
              id="quiet-grace-seconds"
              min={0}
              name="quietGraceSeconds"
              type="number"
            />
          </div>
        </form>
      </section>
    </main>
  );
}
