'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import SignupForm from '@/components/auth/SignupForm';

export default function SignupPage() {
    const [config, setConfig] = useState<{ signupEnabled?: boolean; error?: string } | null>(null);

    useEffect(() => {
        fetch('/api/config')
            .then((r) => r.json())
            .then(setConfig)
            .catch(() => setConfig({ signupEnabled: false }));
    }, []);

    if (!config) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-slate-50 p-4">
                <div className="w-full max-w-sm rounded-xl bg-white p-8 shadow-lg">
                    <p className="text-sm text-slate-500">Loading…</p>
                </div>
            </div>
        );
    }

    if (!config.signupEnabled) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-slate-50 p-4">
                <div className="w-full max-w-sm rounded-xl bg-white p-8 shadow-lg text-center">
                    <h1 className="text-2xl font-semibold text-slate-900">Registration Closed</h1>
                    <p className="mt-2 text-sm text-slate-500">
                        Signup is currently disabled. Contact your administrator.
                    </p>
                    {config.error && <p className="mt-2 text-xs text-red-500">{config.error}</p>}
                    <Link href="/login" className="mt-4 inline-block text-sm text-blue-600 hover:underline">
                        Back to login
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="flex min-h-screen items-center justify-center bg-slate-50 p-4">
            <div className="w-full max-w-sm rounded-xl bg-white p-8 shadow-lg">
                <h1 className="text-2xl font-semibold text-slate-900">Create Account</h1>
                <p className="mt-1 text-sm text-slate-500">Register for your organization console</p>
                <div className="mt-6">
                    <SignupForm />
                </div>
                <p className="mt-4 text-center text-sm text-slate-500">
                    Already have an account?{' '}
                    <Link href="/login" className="text-blue-600 hover:underline">
                        Sign in
                    </Link>
                </p>
            </div>
        </div>
    );
}
