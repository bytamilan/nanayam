import { NextRequest } from 'next/server';

const mockFetch = jest.fn();
global.fetch = mockFetch as unknown as typeof fetch;

function requestWithToken(token?: string): NextRequest {
  return {
    cookies: { get: (name: string) => (token && name === 'nanayam_token' ? { value: token } : undefined) },
  } as unknown as NextRequest;
}

describe('/api/auth/me', () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  it('returns 401 without calling the gateway when no session cookie is present', async () => {
    const { GET } = await import('./route');

    const res = await GET(requestWithToken());

    expect(res.status).toBe(401);
    expect(mockFetch).not.toHaveBeenCalled();
  });

  it('forwards the session cookie as a bearer token', async () => {
    const { GET } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      status: 200,
      json: async () => ({ username: 'admin', org: 'ACBMSP', role: 'admin' }),
    });

    const res = await GET(requestWithToken('jwt-token-value'));
    const body = await res.json();

    expect(res.status).toBe(200);
    expect(body.username).toBe('admin');

    const [, init] = mockFetch.mock.calls[0];
    expect(init.headers.Authorization).toBe('Bearer jwt-token-value');
  });

  it('passes through a gateway rejection of an expired token', async () => {
    const { GET } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      status: 401,
      json: async () => ({ error: 'invalid token' }),
    });

    const res = await GET(requestWithToken('expired-token'));

    expect(res.status).toBe(401);
  });

  it('reports a 500 when the gateway is unreachable', async () => {
    const { GET } = await import('./route');
    mockFetch.mockRejectedValueOnce(new Error('ECONNREFUSED'));

    const res = await GET(requestWithToken('jwt-token-value'));
    const body = await res.json();

    expect(res.status).toBe(500);
    expect(body.error).toContain('ECONNREFUSED');
  });
});
