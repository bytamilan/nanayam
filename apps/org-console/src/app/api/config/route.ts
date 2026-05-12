import { NextResponse } from 'next/server';

const GATEWAY_URL = process.env.GATEWAY_URL || 'http://localhost:8080';

export async function GET() {
    try {
        const res = await fetch(`${GATEWAY_URL}/v1/Config`, { cache: 'no-store' });
        const text = await res.text();
        let data;
        try {
            data = JSON.parse(text);
        } catch {
            return NextResponse.json(
                { signupEnabled: false, raw: text.slice(0, 200) },
                { status: 502 }
            );
        }
        return NextResponse.json(data, { status: res.status });
    } catch (err: any) {
        return NextResponse.json({ signupEnabled: false, error: err.message }, { status: 500 });
    }
}
