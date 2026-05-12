import { POST } from './route';
import { NextRequest } from 'next/server';

const mockFetch = jest.fn();
global.fetch = mockFetch as any;

function mockReq(body: object): NextRequest {
  return {
    cookies: { get: () => undefined },
    json: async () => body,
  } as unknown as NextRequest;
}

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

    const res = await POST(mockReq({ complaintId: 'COMP-001', action: 'acknowledge' }));
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

    const res = await POST(mockReq({}));
    const body = await res.json();

    expect(res.status).toBe(502);
    expect(body.success).toBe(false);
    expect(body.error).toContain('Gateway error');
  });
});
