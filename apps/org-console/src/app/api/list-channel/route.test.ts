import { GET } from './route';

describe('/api/list-channel', () => {
  it('returns default channel', async () => {
    delete process.env.FABRIC_CHANNEL;
    const res = await GET();
    const body = await res.json();
    expect(body.channels).toEqual(['complaint-channel']);
  });

  it('returns configured channel from env', async () => {
    process.env.FABRIC_CHANNEL = 'mychannel';
    const res = await GET();
    const body = await res.json();
    expect(body.channels).toEqual(['mychannel']);
    delete process.env.FABRIC_CHANNEL;
  });
});
