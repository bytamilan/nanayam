export async function listChannels(): Promise<string[]> {
    const res = await fetch('/api/list-channel');
    const { assetIds } = await res.json();
    return assetIds;
}