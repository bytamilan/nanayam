import { NextResponse } from 'next/server';

export async function POST(request: Request) {
    const gateway = process.env.GATEWAY_URL || 'http://localhost:8080';
    try {
        const body = await request.json();
        const res = await fetch(`${gateway}/v1/UpdateComplaint`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        });
        const json = await res.json();
        return NextResponse.json(json);
    } catch (err: any) {
        return NextResponse.json({ success: false, error: err.message || 'Failed to update complaint' }, { status: 502 });
    }
}
