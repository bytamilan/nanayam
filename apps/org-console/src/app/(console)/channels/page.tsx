import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';

export const dynamic = 'force-dynamic';

interface OrgInfo {
    mspId: string;
    name: string;
    peers: string[];
    role: string;
}

interface ChannelInfo {
    channel: string;
    chaincode: string;
    mspId: string;
    organizations: OrgInfo[];
    orderers: string[];
}

async function getChannelInfo(): Promise<ChannelInfo | null> {
    const cookieStore = await cookies();
    const token = cookieStore.get('nanayam_token')?.value;
    if (!token) return null;

    const res = await fetch(`${process.env.GATEWAY_URL || 'http://localhost:8080'}/v1/ChannelInfo`, {
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store',
    });
    if (!res.ok) return null;
    return res.json();
}

function orgColor(mspId: string): string {
    switch (mspId) {
        case 'ACBMSP': return 'border-blue-200 bg-blue-50 text-blue-900';
        case 'DeptMSP': return 'border-yellow-200 bg-yellow-50 text-yellow-900';
        case 'OversightMSP': return 'border-green-200 bg-green-50 text-green-900';
        case 'JudiciaryMSP': return 'border-purple-200 bg-purple-50 text-purple-900';
        default: return 'border-slate-200 bg-slate-50 text-slate-900';
    }
}

function orgBadge(mspId: string): string {
    switch (mspId) {
        case 'ACBMSP': return 'bg-blue-600';
        case 'DeptMSP': return 'bg-yellow-600';
        case 'OversightMSP': return 'bg-green-600';
        case 'JudiciaryMSP': return 'bg-purple-600';
        default: return 'bg-slate-600';
    }
}

export default async function ChannelsPage() {
    const info = await getChannelInfo();
    if (!info) redirect('/login');

    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-2xl font-semibold text-slate-900">Channels</h1>
                <p className="text-sm text-slate-500 mt-1">
                    Network topology and channel configuration.
                </p>
            </div>

            <div className="rounded-xl border border-slate-200 bg-white p-6">
                <h2 className="text-sm font-semibold text-slate-900 uppercase tracking-wide">Channel</h2>
                <div className="mt-3 flex items-center gap-4">
                    <div className="rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white">
                        {info.channel}
                    </div>
                    <div className="text-sm text-slate-600">
                        Chaincode: <span className="font-medium text-slate-900">{info.chaincode}</span>
                    </div>
                    <div className="text-sm text-slate-600">
                        Connected as: <span className="font-medium text-slate-900">{info.mspId}</span>
                    </div>
                </div>
            </div>

            <div>
                <h2 className="text-sm font-semibold text-slate-900 uppercase tracking-wide mb-3">Organizations</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {info.organizations.map((org) => (
                        <div
                            key={org.mspId}
                            className={`rounded-xl border p-5 ${orgColor(org.mspId)}`}
                        >
                            <div className="flex items-center gap-3 mb-3">
                                <span className={`inline-block h-3 w-3 rounded-full ${orgBadge(org.mspId)}`} />
                                <h3 className="font-semibold">{org.name}</h3>
                                <span className="text-xs opacity-70 font-mono">{org.mspId}</span>
                            </div>
                            <p className="text-sm opacity-90 mb-3">{org.role}</p>
                            <div className="space-y-1">
                                <p className="text-xs font-medium opacity-70">Peers</p>
                                {org.peers.map((peer) => (
                                    <code key={peer} className="block text-xs font-mono bg-white/60 rounded px-2 py-1">
                                        {peer}
                                    </code>
                                ))}
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            <div className="rounded-xl border border-slate-200 bg-white p-6">
                <h2 className="text-sm font-semibold text-slate-900 uppercase tracking-wide">Orderers</h2>
                <div className="mt-3 space-y-2">
                    {info.orderers.map((o) => (
                        <code key={o} className="block text-sm font-mono bg-slate-50 rounded px-3 py-2 border border-slate-100">
                            {o}
                        </code>
                    ))}
                </div>
            </div>
        </div>
    );
}
