import { NextResponse } from 'next/server';

export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
    const gateway = process.env.GATEWAY_URL || 'http://localhost:8080';
    try {
        const { id } = await params;
        const res = await fetch(`${gateway}/v1/GetComplaintHistory?complaintId=${id}`);
        if (!res.ok) {
            return NextResponse.json({ error: `Gateway error: ${res.status}` }, { status: 502 });
        }
        const json = await res.json();
        return NextResponse.json(json);
    } catch (err: any) {
        return NextResponse.json({ error: err.message || 'Failed to fetch history' }, { status: 502 });
    }
}
