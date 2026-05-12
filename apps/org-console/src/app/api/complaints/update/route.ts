import { NextRequest, NextResponse } from 'next/server';

const GATEWAY_URL = process.env.GATEWAY_URL || 'http://localhost:8080';

function getAuthHeader(req: NextRequest): HeadersInit {
    const token = req.cookies.get('nanayam_token')?.value;
    return token ? { Authorization: `Bearer ${token}` } : {};
}

export async function POST(req: NextRequest) {
    try {
        const body = await req.json();
        const headers = getAuthHeader(req);
        const res = await fetch(`${GATEWAY_URL}/v1/UpdateComplaint`, {
            method: 'POST',
            headers: { ...headers, 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });
        const json = await res.json();
        return NextResponse.json(json, { status: res.status });
    } catch (err: any) {
        return NextResponse.json({ success: false, error: err.message || 'Failed to update complaint' }, { status: 502 });
    }
}
