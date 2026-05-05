import { GET } from './route';

const mockFetch = jest.fn();
global.fetch = mockFetch as any;

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

    const res = await GET();
    const body = await res.json();

    expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/v1/ListComplaints');
    expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/v1/QueryComplaint?complaintId=COMP-001');
    expect(body.complaints).toHaveLength(1);
    expect(body.complaints[0].complaintId).toBe('COMP-001');
  });

  it('uses GATEWAY_URL env var when set', async () => {
    process.env.GATEWAY_URL = 'http://gateway:8080';
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ complaintIds: [] }),
    });

    const res = await GET();
    await res.json();

    expect(mockFetch).toHaveBeenCalledWith('http://gateway:8080/v1/ListComplaints');
  });

  it('returns empty list with error when gateway is down', async () => {
    mockFetch.mockRejectedValue(new Error('Connection refused'));

    const res = await GET();
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

    const res = await GET();
    const body = await res.json();

    expect(res.status).toBe(502);
    expect(body.complaints).toEqual([]);
    expect(body.error).toContain('Gateway error: 500');
  });
});
