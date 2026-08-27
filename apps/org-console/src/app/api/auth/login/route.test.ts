import { NextRequest } from 'next/server';

const mockFetch = jest.fn();
global.fetch = mockFetch as unknown as typeof fetch;

function jsonRequest(body: unknown): NextRequest {
  return {
    json: async () => body,
  } as unknown as NextRequest;
}

describe('/api/auth/login', () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  it('sets an httpOnly session cookie on a successful login', async () => {
    const { POST } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      text: async () => JSON.stringify({ token: 'jwt-token-value' }),
    });

    const res = await POST(jsonRequest({ username: 'admin', password: 'admin' }));

    expect(res.status).toBe(200);
    await expect(res.json()).resolves.toEqual({ success: true });

    const cookie = res.cookies.get('nanayam_token');
    expect(cookie?.value).toBe('jwt-token-value');
    // The token must be unreadable from client-side JavaScript, otherwise an
    // XSS on the console hands over a ledger session.
    expect(cookie?.httpOnly).toBe(true);
    expect(cookie?.sameSite).toBe('lax');
    expect(cookie?.path).toBe('/');
  });

  it('never puts the raw token in the response body', async () => {
    const { POST } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      text: async () => JSON.stringify({ token: 'jwt-token-value' }),
    });

    const res = await POST(jsonRequest({ username: 'admin', password: 'admin' }));
    const body = await res.json();

    expect(JSON.stringify(body)).not.toContain('jwt-token-value');
  });

  it('passes through the gateway status when credentials are rejected', async () => {
    const { POST } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      text: async () => JSON.stringify({ error: 'invalid username or password' }),
    });

    const res = await POST(jsonRequest({ username: 'admin', password: 'wrong' }));

    expect(res.status).toBe(401);
    expect(res.cookies.get('nanayam_token')).toBeUndefined();
  });

  it('rejects a 200 response that carries no token', async () => {
    const { POST } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      text: async () => JSON.stringify({ unexpected: true }),
    });

    const res = await POST(jsonRequest({ username: 'admin', password: 'admin' }));

    expect(res.cookies.get('nanayam_token')).toBeUndefined();
  });

  it('reports a 502 when the gateway returns non-JSON', async () => {
    const { POST } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      text: async () => '<html>502 Bad Gateway</html>',
    });

    const res = await POST(jsonRequest({ username: 'admin', password: 'admin' }));
    const body = await res.json();

    expect(res.status).toBe(502);
    expect(body.error).toContain('non-JSON');
  });

  it('reports a 500 when the gateway is unreachable', async () => {
    const { POST } = await import('./route');
    mockFetch.mockRejectedValueOnce(new Error('ECONNREFUSED'));

    const res = await POST(jsonRequest({ username: 'admin', password: 'admin' }));
    const body = await res.json();

    expect(res.status).toBe(500);
    expect(body.error).toContain('ECONNREFUSED');
  });
});
