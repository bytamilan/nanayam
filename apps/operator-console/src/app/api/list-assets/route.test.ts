import { GET } from './route';

const mockFetch = jest.fn();
global.fetch = mockFetch as any;

describe('/api/list-assets', () => {
  beforeEach(() => {
    mockFetch.mockClear();
    delete process.env.GATEWAY_URL;
  });

  it('returns asset ids from gateway', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ assetIds: ['asset1', 'asset2'] }),
    });

    const res = await GET();
    const body = await res.json();

    expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/v1/ListAssets');
    expect(body.assetIds).toEqual(['asset1', 'asset2']);
  });

  it('returns error when gateway fails', async () => {
    mockFetch.mockRejectedValue(new Error('Gateway down'));

    const res = await GET();
    const body = await res.json();

    expect(res.status).toBe(502);
    expect(body.assetIds).toEqual([]);
    expect(body.error).toContain('Gateway down');
  });

  it('returns empty list when chaincode function is missing', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      text: async () => 'rpc error: code = Unknown desc = Function GetAllAssets not found in contract SmartContract',
    });

    const res = await GET();
    const body = await res.json();

    expect(res.status).toBe(200);
    expect(body.assetIds).toEqual([]);
  });
});
