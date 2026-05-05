import { NextResponse } from 'next/server';

export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
    const { id } = await params;
    const res = await fetch(`http://gateway:8080/v1/GetComplaintHistory?complaintId=${id}`);
    const json = await res.json();
    return NextResponse.json(json);
}
