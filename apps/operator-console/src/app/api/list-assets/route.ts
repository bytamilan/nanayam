import { NextResponse } from 'next/server';

export async function GET() {
    const gateway = process.env.GATEWAY_URL || 'http://localhost:8080';
    try {
        const res = await fetch(`${gateway}/v1/ListAssets`);
        if (!res.ok) {
            const raw = await res.text();
            const lower = raw.toLowerCase();
            if (lower.includes('function') && lower.includes('not found')) {
                return NextResponse.json({ assetIds: [] }, { status: 200 });
            }
            return NextResponse.json(
                { assetIds: [], error: `Gateway error: ${res.status}${raw ? ` - ${raw}` : ''}` },
                { status: 502 }
            );
        }
        const json = await res.json();
        return NextResponse.json(json);
    } catch (err: any) {
        return NextResponse.json({ assetIds: [], error: err.message || 'Failed to fetch assets' }, { status: 502 });
    }
}
