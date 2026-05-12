import { NextRequest, NextResponse } from 'next/server';

const GATEWAY_URL = process.env.GATEWAY_URL || 'http://localhost:8080';

export async function GET(req: NextRequest) {
    try {
        const token = req.cookies.get('nanayam_token')?.value;
        if (!token) {
            return NextResponse.json({ error: 'unauthorized' }, { status: 401 });
        }
        const res = await fetch(`${GATEWAY_URL}/v1/LedgerActivity`, {
            method: 'GET',
            headers: { Authorization: `Bearer ${token}` },
        });
        const data = await res.json();
        return NextResponse.json(data, { status: res.status });
    } catch (err: any) {
        return NextResponse.json({ error: err.message }, { status: 500 });
    }
}
