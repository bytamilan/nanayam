import { NextResponse } from 'next/server';

export async function GET() {
    const channel = process.env.FABRIC_CHANNEL || 'complaint-channel';
    return NextResponse.json({ channels: [channel] });
}
