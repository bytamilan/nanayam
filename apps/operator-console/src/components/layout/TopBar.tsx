'use client';

import { useRouter } from 'next/navigation';
import { useAuth } from '@/components/auth/AuthProvider';
import { orgBgClass } from '@/lib/auth';

export default function TopBar() {
    const router = useRouter();
    const { user, logout } = useAuth();

    const handleLogout = async () => {
        await logout();
        router.push('/login');
    };

    return (
        <header className="flex items-center justify-between border-b border-slate-200 bg-white px-6 py-3">
            <h1 className="text-sm font-medium text-slate-500">
                Organization Console
            </h1>
            <div className="flex items-center gap-4">
                {user && (
                    <>
                        <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium text-white ${orgBgClass(user.org)}`}>
                            {user.org}
                        </span>
                        <span className="text-sm text-slate-700">{user.username}</span>
                    </>
                )}
                <button
                    onClick={handleLogout}
                    className="text-sm text-slate-500 hover:text-slate-900"
                >
                    Log out
                </button>
            </div>
        </header>
    );
}
