import { ArrowRight, ShieldCheck } from 'lucide-react';

export function LoginPage() {
  return (
    <main className="login-shell">
      <section className="login-panel" aria-labelledby="login-title">
        <div className="login-panel__masthead">
          <div className="login-panel__mark" aria-hidden="true">
            <ShieldCheck size={18} strokeWidth={2.1} />
          </div>
          <div>
            <p className="login-panel__eyebrow">Disk management console</p>
            <h1 id="login-title">DiskHM</h1>
          </div>
        </div>

        <div className="login-panel__copy">
          <p>Authenticate with the local access token to manage disk sleep policy, topology, and events.</p>
        </div>

        <form className="login-form">
          <label className="field">
            <span className="field__label">Token</span>
            <input
              className="field__input"
              name="token"
              type="password"
              autoComplete="current-password"
              placeholder="Enter local access token"
            />
          </label>

          <button className="login-form__submit" type="submit">
            <span>Sign in</span>
            <ArrowRight size={16} strokeWidth={2.2} />
          </button>
        </form>
      </section>

      <aside className="login-aside" aria-label="Status summary">
        <div className="status-card">
          <p className="status-card__label">Access scope</p>
          <p className="status-card__value">Local admin session</p>
          <p className="status-card__meta">Loopback-only by default. Mutating requests require a CSRF token after sign-in.</p>
        </div>
        <div className="status-card">
          <p className="status-card__label">Telemetry mode</p>
          <p className="status-card__value">Safe refresh preferred</p>
          <p className="status-card__meta">Topology, events, and settings are scaffolded for the first shell pass.</p>
        </div>
      </aside>
    </main>
  );
}
