export async function listChannels(): Promise<string[]> {
    const res = await fetch('/api/list-channel');
    const json = await res.json();
    return json.channels || [];
}
