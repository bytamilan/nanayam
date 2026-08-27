const mockFetch = jest.fn();
global.fetch = mockFetch as unknown as typeof fetch;

describe('/api/config', () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  it('returns the gateway configuration', async () => {
    const { GET } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      text: async () => JSON.stringify({ signupEnabled: true, channel: 'mychannel', chaincode: 'basic' }),
    });

    const res = await GET();
    const body = await res.json();

    expect(res.status).toBe(200);
    expect(body).toEqual({ signupEnabled: true, channel: 'mychannel', chaincode: 'basic' });
  });

  it('requests fresh config rather than a cached copy', async () => {
    const { GET } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      text: async () => JSON.stringify({ signupEnabled: false }),
    });

    await GET();

    const [, init] = mockFetch.mock.calls[0];
    expect(init.cache).toBe('no-store');
  });

  it('falls back to signup disabled when the gateway returns non-JSON', async () => {
    const { GET } = await import('./route');
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 502,
      text: async () => '<html>Bad Gateway</html>',
    });

    const res = await GET();
    const body = await res.json();

    expect(res.status).toBe(502);
    // Failing closed matters here: a parse failure must not present the signup
    // form as if registration were open.
    expect(body.signupEnabled).toBe(false);
  });

  it('falls back to signup disabled when the gateway is unreachable', async () => {
    const { GET } = await import('./route');
    mockFetch.mockRejectedValueOnce(new Error('ECONNREFUSED'));

    const res = await GET();
    const body = await res.json();

    expect(res.status).toBe(500);
    expect(body.signupEnabled).toBe(false);
  });
});
