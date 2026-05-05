export async function listAssets(): Promise<string[]> {
    const res = await fetch('/api/list-assets');
    const { assetIds } = await res.json();
    return assetIds;
}