import { NextResponse } from 'next/server';

export async function GET() {
    const gateway = process.env.GATEWAY_URL || 'http://localhost:8080';
    try {
        const res = await fetch(`${gateway}/v1/ListAssets`);
        if (!res.ok) {
            return NextResponse.json({ assetIds: [], error: `Gateway error: ${res.status}` }, { status: 502 });
        }
        const json = await res.json();
        return NextResponse.json(json);
    } catch (err: any) {
        return NextResponse.json({ assetIds: [], error: err.message || 'Failed to fetch assets' }, { status: 502 });
    }
}
