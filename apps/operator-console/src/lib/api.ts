export interface Asset { assetId: string; color: string; size: number; }

export async function listAssets(): Promise<string[]> {
    const res = await fetch('/api/ListAssets');
    const json = await res.json();
    return json.assetIds;
}

export async function createAsset(asset: Asset): Promise<boolean> {
    const res = await fetch('/api/CreateAsset', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(asset)
    });
    const json = await res.json();
    return json.success;
}