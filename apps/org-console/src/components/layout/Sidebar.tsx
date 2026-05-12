'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuth } from '@/components/auth/AuthProvider';
import { orgBorderClass, orgBgClass } from '@/lib/auth';

const NAV = [
    { href: '/dashboard', label: 'Dashboard', icon: DashboardIcon },
    { href: '/complaints', label: 'Complaints', icon: ComplaintIcon },
    { href: '/channels', label: 'Channels', icon: ChannelIcon },
    { href: '/ledger', label: 'Ledger', icon: LedgerIcon },
];

export default function Sidebar() {
    const pathname = usePathname();
    const { user } = useAuth();
    const org = user?.org || '';

    return (
        <aside className="flex w-64 flex-col border-r border-slate-200 bg-white">
            <div className="flex items-center gap-3 px-6 py-5 border-b border-slate-100">
                <div className={`h-8 w-8 rounded-lg ${orgBgClass(org)} flex items-center justify-center text-white font-bold text-sm`}>
                    N
                </div>
                <div>
                    <h2 className="text-sm font-semibold text-slate-900">Nanayam</h2>
                    <p className="text-xs text-slate-500">{user?.org || 'Console'}</p>
                </div>
            </div>
            <nav className="flex-1 px-3 py-4 space-y-1">
                {NAV.map((item) => {
                    const active = pathname === item.href || pathname.startsWith(item.href + '/');
                    return (
                        <Link
                            key={item.href}
                            href={item.href}
                            className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                                active
                                    ? `bg-slate-50 text-slate-900 ${orgBorderClass(org)} border-l-4`
                                    : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
                            }`}
                        >
                            <item.icon className="h-5 w-5" active={active} />
                            {item.label}
                        </Link>
                    );
                })}
            </nav>
        </aside>
    );
}

function DashboardIcon({ className, active }: { className?: string; active?: boolean }) {
    return (
        <svg className={className} fill="none" stroke={active ? 'currentColor' : '#64748b'} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
        </svg>
    );
}

function ComplaintIcon({ className, active }: { className?: string; active?: boolean }) {
    return (
        <svg className={className} fill="none" stroke={active ? 'currentColor' : '#64748b'} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
    );
}

function ChannelIcon({ className, active }: { className?: string; active?: boolean }) {
    return (
        <svg className={className} fill="none" stroke={active ? 'currentColor' : '#64748b'} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
    );
}

function LedgerIcon({ className, active }: { className?: string; active?: boolean }) {
    return (
        <svg className={className} fill="none" stroke={active ? 'currentColor' : '#64748b'} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
    );
}
