'use client';

import { ReactNode, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/components/auth/AuthProvider';
import ConsoleLayout from '@/components/layout/ConsoleLayout';

export default function ProtectedLayout({ children }: { children: ReactNode }) {
    const router = useRouter();
    const { user, loading } = useAuth();

    useEffect(() => {
        if (!loading && !user) {
            router.push('/login');
        }
    }, [loading, user, router]);

    if (loading) {
        return (
            <div className="flex h-screen items-center justify-center bg-slate-50">
                <div className="text-slate-500">Loading…</div>
            </div>
        );
    }

    if (!user) {
        return null;
    }

    return <ConsoleLayout>{children}</ConsoleLayout>;
}
