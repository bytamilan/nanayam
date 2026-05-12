import { NextRequest, NextResponse } from 'next/server';

const GATEWAY_URL = process.env.GATEWAY_URL || 'http://localhost:8080';

function getAuthHeader(req: NextRequest): HeadersInit {
    const token = req.cookies.get('nanayam_token')?.value;
    return token ? { Authorization: `Bearer ${token}` } : {};
}

export async function GET(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
    try {
        const { id } = await params;
        const headers = getAuthHeader(req);
        const res = await fetch(`${GATEWAY_URL}/v1/GetComplaintHistory?complaintId=${id}`, { headers });
        if (!res.ok) {
            return NextResponse.json({ error: `Gateway error: ${res.status}` }, { status: 502 });
        }
        const json = await res.json();
        return NextResponse.json(json);
    } catch (err: any) {
        return NextResponse.json({ error: err.message || 'Failed to fetch history' }, { status: 502 });
    }
}
