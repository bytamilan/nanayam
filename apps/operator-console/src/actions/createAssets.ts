export async function createAsset(data: { assetId: string; color: string; size: number; }): Promise<boolean> {
    const res = await fetch('/api/create-asset', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    });
    const { success } = await res.json();
    return success;
}