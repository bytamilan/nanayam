import { NextResponse } from 'next/server';
export async function POST(request: Request) {
    const body = await request.json();
    const res = await fetch('http://gateway:8080/v1/CreateAsset', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
    });
    const json = await res.json();
    return NextResponse.json(json);
}