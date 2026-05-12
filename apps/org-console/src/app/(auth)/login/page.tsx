import Link from 'next/link';
import LoginForm from '@/components/auth/LoginForm';

export const dynamic = 'force-dynamic';

async function getConfig() {
    try {
        const res = await fetch(`${process.env.GATEWAY_URL || 'http://localhost:8080'}/v1/Config`, {
            cache: 'no-store',
        });
        const text = await res.text();
        return JSON.parse(text);
    } catch {
        return { signupEnabled: false };
    }
}

export default async function LoginPage() {
    const config = await getConfig();

    return (
        <div className="flex min-h-screen items-center justify-center bg-slate-50 p-4">
            <div className="w-full max-w-sm rounded-xl bg-white p-8 shadow-lg">
                <h1 className="text-2xl font-semibold text-slate-900">Nanayam Console</h1>
                <p className="mt-1 text-sm text-slate-500">Sign in to your organization</p>
                <div className="mt-6">
                    <LoginForm />
                </div>
                {config.signupEnabled && (
                    <p className="mt-4 text-center text-sm text-slate-500">
                        Don&apos;t have an account?{' '}
                        <Link href="/signup" className="text-blue-600 hover:underline">
                            Sign up
                        </Link>
                    </p>
                )}
            </div>
        </div>
    );
}
