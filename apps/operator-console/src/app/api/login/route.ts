import { NextResponse } from 'next/server';
import { enrollUser } from '@/lib/bevelHelper';

export async function POST(req: Request) {
    const { userId, password } = await req.json();
    try {
        await enrollUser(userId, password);
        return NextResponse.json({ success: true });
    } catch (error) {
        return NextResponse.json({ success: false, error: String(error) }, { status: 401 });
    }
}