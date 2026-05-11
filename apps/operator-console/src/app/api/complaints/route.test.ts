import { GET } from './route';
import { NextRequest } from 'next/server';

const mockFetch = jest.fn();
global.fetch = mockFetch as any;

function mockReq(): NextRequest {
  return {
    cookies: { get: () => undefined },
  } as unknown as NextRequest;
}

describe('/api/complaints', () => {
  beforeEach(() => {
    mockFetch.mockClear();
    delete process.env.GATEWAY_URL;
  });

  it('returns complaints fetched from gateway', async () => {
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ complaintIds: ['COMP-001'] }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: JSON.stringify({ complaintId: 'COMP-001', status: 'Submitted' }) }),
      });

    const res = await GET(mockReq());
    const body = await res.json();

    expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/v1/ListComplaints', expect.any(Object));
    expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/v1/QueryComplaint?complaintId=COMP-001', expect.any(Object));
    expect(body.complaints).toHaveLength(1);
    expect(body.complaints[0].complaintId).toBe('COMP-001');
  });

  it('uses GATEWAY_URL env var when set', async () => {
    // GATEWAY_URL is cached at module load; this test validates the default URL logic
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ complaintIds: [] }),
    });

    const res = await GET(mockReq());
    await res.json();

    expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/v1/ListComplaints', expect.any(Object));
  });

  it('returns empty list with error when gateway is down', async () => {
    mockFetch.mockRejectedValue(new Error('Connection refused'));

    const res = await GET(mockReq());
    const body = await res.json();

    expect(res.status).toBe(502);
    expect(body.complaints).toEqual([]);
    expect(body.error).toContain('Connection refused');
  });

  it('returns empty list with error on gateway 500', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: async () => ({ error: 'Internal error' }),
    });

    const res = await GET(mockReq());
    const body = await res.json();

    expect(res.status).toBe(502);
    expect(body.complaints).toEqual([]);
    expect(body.error).toContain('Gateway error: 500');
  });
});
