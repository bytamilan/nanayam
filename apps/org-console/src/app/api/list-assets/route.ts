import { NextRequest, NextResponse } from 'next/server';

const GATEWAY_URL = process.env.GATEWAY_URL || 'http://localhost:8080';

function getAuthHeader(req: NextRequest): HeadersInit {
    const token = req.cookies.get('nanayam_token')?.value;
    return token ? { Authorization: `Bearer ${token}` } : {};
}

export async function GET(req: NextRequest) {
    try {
        const headers = getAuthHeader(req);
        const res = await fetch(`${GATEWAY_URL}/v1/ListAssets`, { headers });
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
