import { POST } from './route';

const mockFetch = jest.fn();
global.fetch = mockFetch as any;

describe('/api/complaints/submit', () => {
  beforeEach(() => {
    mockFetch.mockClear();
    delete process.env.GATEWAY_URL;
  });

  it('forwards submit to gateway', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ success: true }),
    });

    const req = new Request('http://localhost/api/complaints/submit', {
      method: 'POST',
      body: JSON.stringify({ complaintId: 'COMP-001', category: 'bribery' }),
    });

    const res = await POST(req);
    const body = await res.json();

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/v1/SubmitComplaint',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ complaintId: 'COMP-001', category: 'bribery' }),
      })
    );
    expect(body.success).toBe(true);
  });

  it('returns error when gateway is unreachable', async () => {
    mockFetch.mockRejectedValue(new Error('Connection refused'));

    const req = new Request('http://localhost/api/complaints/submit', {
      method: 'POST',
      body: JSON.stringify({}),
    });

    const res = await POST(req);
    const body = await res.json();

    expect(res.status).toBe(502);
    expect(body.success).toBe(false);
    expect(body.error).toContain('Connection refused');
  });
});
