import { type FormEvent, useState } from 'react';
import { ArrowRight, ShieldCheck } from 'lucide-react';

import { http, setCsrfToken } from '../../lib/http';

export function LoginPage() {
  const [token, setToken] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!token.trim()) {
      setErrorMessage('Enter the local access token first.');
      setMessage(null);
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);
    setMessage(null);

    try {
      const response = await http.request('/api/session/login', {
        method: 'POST',
        body: { token },
      });

      if (!response.ok) {
        setCsrfToken(null);
        setErrorMessage(response.status === 401 ? 'Access token rejected.' : `Sign-in failed (HTTP ${response.status}).`);
        return;
      }

      setCsrfToken(response.headers.get('X-CSRF-Token'));
      setMessage('Signed in. Local session is ready.');
    } catch {
      setCsrfToken(null);
      setErrorMessage('Sign-in failed because the DiskHM service could not be reached.');
    } finally {
      setIsSubmitting(false);
    }
  }

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

        <form className="login-form" onSubmit={handleSubmit}>
          <label className="field">
            <span className="field__label">Token</span>
            <input
              aria-describedby="login-feedback"
              className="field__input"
              name="token"
              type="password"
              autoComplete="current-password"
              placeholder="Enter local access token"
              value={token}
              onChange={(event) => setToken(event.target.value)}
            />
          </label>

          <button className="login-form__submit" type="submit" disabled={isSubmitting}>
            <span>{isSubmitting ? 'Signing in...' : 'Sign in'}</span>
            <ArrowRight size={16} strokeWidth={2.2} />
          </button>

          <div id="login-feedback" aria-live="polite">
            {message ? <p>{message}</p> : null}
            {errorMessage ? <p role="alert">{errorMessage}</p> : null}
          </div>
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
