import { NextResponse } from 'next/server';
export async function POST(request: Request) {
    const body = await request.json();
    const gateway = process.env.GATEWAY_URL || 'http://localhost:8080';
    const res = await fetch(`${gateway}/v1/CreateAsset`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
    });
    const json = await res.json();
    return NextResponse.json(json);
}