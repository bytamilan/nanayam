describe('/api/auth/logout', () => {
  it('clears the session cookie', async () => {
    const { POST } = await import('./route');

    const res = await POST();

    expect(res.status).toBe(200);
    await expect(res.json()).resolves.toEqual({ success: true });

    const cookie = res.cookies.get('nanayam_token');
    expect(cookie?.value).toBe('');
    // maxAge 0 is what actually expires the cookie in the browser; an empty
    // value alone would leave a live cookie behind.
    expect(cookie?.maxAge).toBe(0);
    expect(cookie?.path).toBe('/');
    expect(cookie?.httpOnly).toBe(true);
  });
});
