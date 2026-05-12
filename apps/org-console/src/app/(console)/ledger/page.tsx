'use client';

import { useEffect, useState } from 'react';

interface BlockSummary {
    number: number;
    hash: string;
    prevHash: string;
    txCount: number;
    timestamp: string;
    dataHash: string;
}

interface LedgerActivity {
    height: number;
    complaintCount: number;
    channel: string;
    chaincode: string;
}

export default function LedgerPage() {
    const [activity, setActivity] = useState<LedgerActivity | null>(null);
    const [blocks, setBlocks] = useState<BlockSummary[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [pageStart, setPageStart] = useState(0);
    const pageSize = 10;

    const fetchData = async () => {
        setLoading(true);
        setError(null);
        try {
            const [activityRes, blocksRes] = await Promise.all([
                fetch('/api/ledger/activity', { credentials: 'include' }),
                fetch(`/api/ledger/blocks?start=${pageStart}&end=${pageStart + pageSize - 1}`, { credentials: 'include' }),
            ]);
            if (!activityRes.ok || !blocksRes.ok) {
                throw new Error('Failed to fetch ledger data');
            }
            const activityData = await activityRes.json();
            const blocksData = await blocksRes.json();
            setActivity(activityData);
            setBlocks(blocksData.blocks || []);
        } catch (err: any) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
    }, [pageStart]);

    const truncate = (s: string, n = 16) => (s.length > n ? s.slice(0, n) + '…' : s);

    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-2xl font-semibold text-slate-900">Ledger</h1>
                <p className="text-sm text-slate-500 mt-1">
                    Blockchain blocks and network activity.
                </p>
            </div>

            {error && (
                <div className="rounded-lg bg-red-50 text-red-700 px-4 py-3 text-sm">{error}</div>
            )}

            {activity && (
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                    <div className="rounded-xl border border-slate-200 bg-white p-5">
                        <p className="text-xs font-medium text-slate-500 uppercase">Block Height</p>
                        <p className="mt-1 text-2xl font-semibold text-slate-900">{activity.height}</p>
                    </div>
                    <div className="rounded-xl border border-slate-200 bg-white p-5">
                        <p className="text-xs font-medium text-slate-500 uppercase">Complaints</p>
                        <p className="mt-1 text-2xl font-semibold text-slate-900">{activity.complaintCount}</p>
                    </div>
                    <div className="rounded-xl border border-slate-200 bg-white p-5">
                        <p className="text-xs font-medium text-slate-500 uppercase">Channel</p>
                        <p className="mt-1 text-lg font-semibold text-slate-900">{activity.channel}</p>
                        <p className="text-xs text-slate-500">{activity.chaincode}</p>
                    </div>
                </div>
            )}

            <div className="rounded-xl border border-slate-200 bg-white overflow-hidden">
                <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
                    <h2 className="text-sm font-semibold text-slate-900">Blocks</h2>
                    <div className="flex gap-2">
                        <button
                            onClick={() => setPageStart((p) => Math.max(0, p - pageSize))}
                            disabled={pageStart === 0 || loading}
                            className="px-3 py-1 text-xs rounded border border-slate-200 hover:bg-slate-50 disabled:opacity-40"
                        >
                            Previous
                        </button>
                        <button
                            onClick={() => setPageStart((p) => p + pageSize)}
                            disabled={blocks.length < pageSize || loading}
                            className="px-3 py-1 text-xs rounded border border-slate-200 hover:bg-slate-50 disabled:opacity-40"
                        >
                            Next
                        </button>
                    </div>
                </div>
                <div className="overflow-x-auto">
                    <table className="min-w-full text-sm">
                        <thead className="bg-slate-50 text-slate-600">
                            <tr>
                                <th className="px-6 py-3 text-left font-medium">#</th>
                                <th className="px-6 py-3 text-left font-medium">Hash</th>
                                <th className="px-6 py-3 text-left font-medium">Prev Hash</th>
                                <th className="px-6 py-3 text-left font-medium">Tx Count</th>
                                <th className="px-6 py-3 text-left font-medium">Data Hash</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100">
                            {blocks.map((b) => (
                                <tr key={b.number} className="hover:bg-slate-50">
                                    <td className="px-6 py-3 font-mono text-slate-900">{b.number}</td>
                                    <td className="px-6 py-3 font-mono text-slate-600">{truncate(b.hash, 20)}</td>
                                    <td className="px-6 py-3 font-mono text-slate-600">{truncate(b.prevHash, 20)}</td>
                                    <td className="px-6 py-3 text-slate-900">{b.txCount}</td>
                                    <td className="px-6 py-3 font-mono text-slate-600">{truncate(b.dataHash, 20)}</td>
                                </tr>
                            ))}
                            {blocks.length === 0 && !loading && (
                                <tr>
                                    <td colSpan={5} className="px-6 py-8 text-center text-slate-400">
                                        No blocks found.
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
