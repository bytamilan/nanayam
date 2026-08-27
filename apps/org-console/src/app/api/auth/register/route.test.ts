import { NextRequest } from 'next/server';

const mockFetch = jest.fn();
global.fetch = mockFetch as unknown as typeof fetch;

function jsonRequest(body: unknown): NextRequest {
  return { json: async () => body } as unknown as NextRequest;
}

describe('/api/auth/register', () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  it('forwards the registration to the gateway and returns the created user', async () => {
    const { POST } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      text: async () => JSON.stringify({ id: 'usr-1', username: 'alice', org: 'ACBMSP' }),
    });

    const res = await POST(jsonRequest({ username: 'alice', password: 'pw', org: 'ACBMSP' }));
    const body = await res.json();

    expect(res.status).toBe(201);
    expect(body.username).toBe('alice');

    const [, init] = mockFetch.mock.calls[0];
    expect(JSON.parse(init.body)).toEqual({ username: 'alice', password: 'pw', org: 'ACBMSP' });
  });

  it('passes through a 403 when signup is disabled on the gateway', async () => {
    const { POST } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 403,
      text: async () => JSON.stringify({ error: 'registration is disabled' }),
    });

    const res = await POST(jsonRequest({ username: 'mallory', password: 'pw' }));
    const body = await res.json();

    expect(res.status).toBe(403);
    expect(body.error).toContain('disabled');
  });

  it('reports a 502 when the gateway returns non-JSON', async () => {
    const { POST } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 502,
      text: async () => '<html>Bad Gateway</html>',
    });

    const res = await POST(jsonRequest({ username: 'alice', password: 'pw' }));
    const body = await res.json();

    expect(res.status).toBe(502);
    expect(body.error).toContain('non-JSON');
  });

  it('reports a 500 when the gateway is unreachable', async () => {
    const { POST } = await import('./route');
    mockFetch.mockRejectedValueOnce(new Error('ECONNREFUSED'));

    const res = await POST(jsonRequest({ username: 'alice', password: 'pw' }));

    expect(res.status).toBe(500);
  });
});
