import {
  listAssets,
  createAsset,
  listComplaints,
  getComplaint,
  submitComplaint,
  updateComplaint,
  getComplaintHistory,
} from './api';

const mockFetch = jest.fn();
global.fetch = mockFetch as unknown as typeof fetch;

function ok(json: unknown) {
  return { ok: true, status: 200, json: async () => json };
}

describe('console API client', () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  it('sends credentials on every request so the session cookie travels', async () => {
    mockFetch.mockResolvedValueOnce(ok({ assetIds: [] }));

    await listAssets();

    const [, init] = mockFetch.mock.calls[0];
    expect(init.credentials).toBe('include');
  });

  it('listAssets returns the ids from the response', async () => {
    mockFetch.mockResolvedValueOnce(ok({ assetIds: ['asset1', 'asset2'] }));

    await expect(listAssets()).resolves.toEqual(['asset1', 'asset2']);
    expect(mockFetch).toHaveBeenCalledWith('/api/list-assets', expect.any(Object));
  });

  it('createAsset posts the asset as JSON', async () => {
    mockFetch.mockResolvedValueOnce(ok({ success: true }));

    await expect(createAsset({ assetId: 'asset1', color: 'blue', size: 5 })).resolves.toBe(true);

    const [url, init] = mockFetch.mock.calls[0];
    expect(url).toBe('/api/create-asset');
    expect(init.method).toBe('POST');
    expect(JSON.parse(init.body)).toEqual({ assetId: 'asset1', color: 'blue', size: 5 });
  });

  it('createAsset reports failure when the route says so', async () => {
    mockFetch.mockResolvedValueOnce(ok({ success: false, error: 'endorsement failed' }));

    await expect(createAsset({ assetId: 'asset1', color: 'blue', size: 5 })).resolves.toBe(false);
  });

  it('listComplaints returns an empty list rather than undefined', async () => {
    mockFetch.mockResolvedValueOnce(ok({}));

    // Callers render this straight into a table, so undefined would throw.
    await expect(listComplaints()).resolves.toEqual([]);
  });

  it('listComplaints returns the complaints when present', async () => {
    const complaints = [{ complaintId: 'COMP-001', status: 'Submitted' }];
    mockFetch.mockResolvedValueOnce(ok({ complaints }));

    await expect(listComplaints()).resolves.toEqual(complaints);
  });

  it('getComplaint returns null on a non-OK response', async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 404, json: async () => ({}) });

    await expect(getComplaint('COMP-404')).resolves.toBeNull();
  });

  it('getComplaint requests the complaint by id', async () => {
    mockFetch.mockResolvedValueOnce(ok({ complaintId: 'COMP-001' }));

    await getComplaint('COMP-001');

    expect(mockFetch).toHaveBeenCalledWith('/api/complaints/COMP-001', expect.any(Object));
  });

  it('submitComplaint posts the complaint payload', async () => {
    mockFetch.mockResolvedValueOnce(ok({ success: true }));

    const payload = {
      complaintId: 'COMP-001',
      category: 'bribery',
      citizenHash: 'sha256:abc',
      descriptionHash: 'sha256:def',
      attachmentsRef: 'ipfs:QmTest',
    };
    await expect(submitComplaint(payload)).resolves.toBe(true);

    const [url, init] = mockFetch.mock.calls[0];
    expect(url).toBe('/api/complaints/submit');
    expect(JSON.parse(init.body)).toEqual(payload);
  });

  it('updateComplaint posts the action', async () => {
    mockFetch.mockResolvedValueOnce(ok({ success: true }));

    await expect(updateComplaint({ complaintId: 'COMP-001', action: 'acknowledge' })).resolves.toBe(true);

    const [url, init] = mockFetch.mock.calls[0];
    expect(url).toBe('/api/complaints/update');
    expect(JSON.parse(init.body)).toEqual({ complaintId: 'COMP-001', action: 'acknowledge' });
  });

  it('getComplaintHistory parses the JSON string the ledger returns', async () => {
    mockFetch.mockResolvedValueOnce(ok({ data: '[{"txId":"tx1"},{"txId":"tx2"}]' }));

    await expect(getComplaintHistory('COMP-001')).resolves.toEqual([{ txId: 'tx1' }, { txId: 'tx2' }]);
  });

  it('getComplaintHistory returns an empty array for malformed history data', async () => {
    mockFetch.mockResolvedValueOnce(ok({ data: 'not json' }));

    await expect(getComplaintHistory('COMP-001')).resolves.toEqual([]);
  });

  it('getComplaintHistory returns an empty array on a non-OK response', async () => {
    mockFetch.mockResolvedValueOnce({ ok: false, status: 500, json: async () => ({}) });

    await expect(getComplaintHistory('COMP-001')).resolves.toEqual([]);
  });
});
