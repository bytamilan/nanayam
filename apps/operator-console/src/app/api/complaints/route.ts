import { NextResponse } from 'next/server';

export async function GET() {
    const gateway = process.env.GATEWAY_URL || 'http://localhost:8080';
    const res = await fetch(`${gateway}/v1/ListComplaints`);
    const json = await res.json();

    // Fetch full complaint details for each ID
    const ids: string[] = json.complaintIds || [];
    const complaints = [];
    for (const id of ids) {
        try {
            const detailRes = await fetch(`http://gateway:8080/v1/QueryComplaint?complaintId=${id}`);
            const detailJson = await detailRes.json();
            complaints.push(JSON.parse(detailJson.data));
        } catch {
            // skip failed lookups
        }
    }

    return NextResponse.json({ complaints });
}
