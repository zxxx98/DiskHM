import { afterEach, describe, expect, it, vi } from 'vitest';

import { http, setCsrfToken } from './http';

describe('http', () => {
  afterEach(() => {
    setCsrfToken(null);
    vi.restoreAllMocks();
  });

  it('preserves urlencoded bodies without forcing a json content type', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(null, {
        status: 204,
      }),
    );
    vi.stubGlobal('fetch', fetchMock);

    await http.request('/api/session/login', {
      method: 'POST',
      body: new URLSearchParams({ token: 'secret' }),
    });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [, init] = fetchMock.mock.calls[0] as [RequestInfo | URL, RequestInit];
    const headers = new Headers(init.headers);

    expect(init.body).toBeInstanceOf(URLSearchParams);
    expect(headers.has('Content-Type')).toBe(false);
  });

  it('serializes json objects and adds the csrf header on mutating requests', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response('{}', {
        status: 200,
        headers: {
          'Content-Type': 'application/json',
        },
      }),
    );
    vi.stubGlobal('fetch', fetchMock);
    setCsrfToken('csrf-token');

    await http.json('/api/disks/example/sleep-after', {
      method: 'POST',
      body: { minutes: 5 },
    });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [, init] = fetchMock.mock.calls[0] as [RequestInfo | URL, RequestInit];
    const headers = new Headers(init.headers);

    expect(init.body).toBe('{"minutes":5}');
    expect(headers.get('Content-Type')).toBe('application/json');
    expect(headers.get('X-CSRF-Token')).toBe('csrf-token');
  });
});
