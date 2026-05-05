import { NextResponse } from 'next/server';
export async function GET() {
    const gateway = process.env.GATEWAY_URL || 'http://localhost:8080';
    const res = await fetch(`${gateway}/v1/ListAssets`);
    const json = await res.json();
    return NextResponse.json(json);
}