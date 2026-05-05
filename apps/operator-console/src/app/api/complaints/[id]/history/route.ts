import { NextResponse } from 'next/server';

export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
    const { id } = await params;
    const gateway = process.env.GATEWAY_URL || 'http://localhost:8080';
    const res = await fetch(`${gateway}/v1/GetComplaintHistory?complaintId=${id}`);
    const json = await res.json();
    return NextResponse.json(json);
}
