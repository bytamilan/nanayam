import { POST } from './route';

const mockFetch = jest.fn();
global.fetch = mockFetch as any;

describe('/api/complaints/update', () => {
  beforeEach(() => {
    mockFetch.mockClear();
    delete process.env.GATEWAY_URL;
  });

  it('forwards update to gateway', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ success: true }),
    });

    const req = new Request('http://localhost/api/complaints/update', {
      method: 'POST',
      body: JSON.stringify({ complaintId: 'COMP-001', action: 'acknowledge' }),
    });

    const res = await POST(req);
    const body = await res.json();

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/v1/UpdateComplaint',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ complaintId: 'COMP-001', action: 'acknowledge' }),
      })
    );
    expect(body.success).toBe(true);
  });

  it('returns error when gateway fails', async () => {
    mockFetch.mockRejectedValue(new Error('Gateway error'));

    const req = new Request('http://localhost/api/complaints/update', {
      method: 'POST',
      body: JSON.stringify({}),
    });

    const res = await POST(req);
    const body = await res.json();

    expect(res.status).toBe(502);
    expect(body.success).toBe(false);
    expect(body.error).toContain('Gateway error');
  });
});
