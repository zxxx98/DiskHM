const CSRF_HEADER_NAME = 'X-CSRF-Token';

let csrfToken: string | null = null;

export function setCsrfToken(token: string | null) {
  csrfToken = token;
}

export function getCsrfToken() {
  return csrfToken;
}

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD';

type JsonPrimitive = boolean | number | string | null;
type JsonValue = JsonPrimitive | JsonValue[] | { [key: string]: JsonValue };
type JsonObject = { [key: string]: JsonValue };
type SupportedBody = FormData | URLSearchParams | string | JsonObject | null | undefined;

type RequestOptions = Omit<RequestInit, 'body' | 'headers' | 'method'> & {
  body?: SupportedBody;
  headers?: HeadersInit;
  method?: HttpMethod;
};

function normalizeBody(body: RequestOptions['body']) {
  if (body == null || body instanceof FormData || body instanceof URLSearchParams || typeof body === 'string') {
    return body;
  }

  return JSON.stringify(body);
}

function isJsonBody(body: RequestOptions['body']): body is JsonObject {
  return body != null && !(body instanceof FormData) && !(body instanceof URLSearchParams) && typeof body !== 'string';
}

function normalizeHeaders(method: HttpMethod, body: RequestOptions['body'], headers?: HeadersInit) {
  const nextHeaders = new Headers(headers);

  if (isJsonBody(body) && !nextHeaders.has('Content-Type')) {
    nextHeaders.set('Content-Type', 'application/json');
  }

  if (method !== 'GET' && method !== 'HEAD' && csrfToken) {
    nextHeaders.set(CSRF_HEADER_NAME, csrfToken);
  }

  return nextHeaders;
}

async function request(input: RequestInfo | URL, options: RequestOptions = {}) {
  const method = options.method ?? 'GET';
  const body = normalizeBody(options.body);
  const headers = normalizeHeaders(method, options.body, options.headers);

  return fetch(input, {
    ...options,
    method,
    body,
    headers,
    credentials: 'include',
  });
}

async function json<T>(input: RequestInfo | URL, options: RequestOptions = {}) {
  const response = await request(input, options);

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }

  if (response.status === 204) {
    return null as T;
  }

  return (await response.json()) as T;
}

export const http = {
  request,
  json,
};
