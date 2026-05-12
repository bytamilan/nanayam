import { NextRequest, NextResponse } from 'next/server';

const GATEWAY_URL = process.env.GATEWAY_URL || 'http://localhost:8080';

function getAuthHeader(req: NextRequest): HeadersInit {
    const token = req.cookies.get('nanayam_token')?.value;
    return token ? { Authorization: `Bearer ${token}` } : {};
}

export async function GET(req: NextRequest) {
    try {
        const headers = getAuthHeader(req);
        const res = await fetch(`${GATEWAY_URL}/v1/ListComplaints`, { headers });
        if (!res.ok) {
            return NextResponse.json({ complaints: [], error: `Gateway error: ${res.status}` }, { status: 502 });
        }
        const json = await res.json();

        const ids: string[] = json.complaintIds || [];
        const complaints = [];
        for (const id of ids) {
            try {
                const detailRes = await fetch(`${GATEWAY_URL}/v1/QueryComplaint?complaintId=${id}`, { headers });
                const detailJson = await detailRes.json();
                complaints.push(JSON.parse(detailJson.data));
            } catch {
                // skip failed lookups
            }
        }

        return NextResponse.json({ complaints });
    } catch (err: any) {
        return NextResponse.json({ complaints: [], error: err.message || 'Failed to fetch complaints' }, { status: 502 });
    }
}
