import { NextResponse } from 'next/server';
export async function GET() {
    const res = await fetch('http://gateway:8080/v1/ListAssets');
    const json = await res.json();
    return NextResponse.json(json);
}