'use client';
import { useEffect, useState } from 'react';
import { listComplaints, Complaint, updateComplaint, getComplaintHistory } from '../lib/api';

const STATUS_COLORS: Record<string, string> = {
    Submitted: 'bg-gray-200 text-gray-800',
    Acknowledged: 'bg-blue-200 text-blue-800',
    UnderInvestigation: 'bg-yellow-200 text-yellow-800',
    ActionTaken: 'bg-purple-200 text-purple-800',
    PendingClosure: 'bg-orange-200 text-orange-800',
    Closed: 'bg-green-200 text-green-800',
    Rejected: 'bg-red-200 text-red-800',
};

function formatDateValue(value: unknown): string {
    if (value == null) return '-';

    if (typeof value === 'string' || typeof value === 'number') {
        const d = new Date(value);
        return Number.isNaN(d.getTime()) ? String(value) : d.toLocaleString();
    }

    if (typeof value === 'object') {
        const ts = value as { seconds?: number | string; nanos?: number | string };
        if (ts.seconds !== undefined) {
            const sec = typeof ts.seconds === 'string' ? Number(ts.seconds) : ts.seconds;
            const nanos = typeof ts.nanos === 'string' ? Number(ts.nanos) : (ts.nanos ?? 0);
            if (Number.isFinite(sec) && Number.isFinite(nanos)) {
                const d = new Date((sec as number) * 1000 + Math.floor((nanos as number) / 1_000_000));
                if (!Number.isNaN(d.getTime())) return d.toLocaleString();
            }
        }
        try {
            return JSON.stringify(value);
        } catch {
            return String(value);
        }
    }

    return String(value);
}

export default function ComplaintTable() {
    const [complaints, setComplaints] = useState<Complaint[]>([]);
    const [selected, setSelected] = useState<Complaint | null>(null);
    const [history, setHistory] = useState<any[]>([]);
    const [loading, setLoading] = useState(false);
    const [message, setMessage] = useState('');

    useEffect(() => { load(); }, []);

    async function load() {
        const data = await listComplaints();
        setComplaints(data);
    }

    async function selectComplaint(c: Complaint) {
        setSelected(c);
        const h = await getComplaintHistory(c.complaintId);
        setHistory(h);
    }

    async function doAction(action: string, value?: string) {
        if (!selected) return;
        setLoading(true);
        setMessage('');
        const ok = await updateComplaint({ complaintId: selected.complaintId, action, value });
        setLoading(false);
        if (ok) {
            setMessage(`Action "${action}" succeeded.`);
            await load();
            const updated = await listComplaints();
            const found = updated.find(x => x.complaintId === selected.complaintId);
            if (found) await selectComplaint(found);
        } else {
            setMessage(`Action "${action}" failed. Check gateway logs.`);
        }
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h2 className="text-xl font-semibold">Complaints</h2>
                <button onClick={load} className="px-3 py-1.5 bg-blue-600 text-white rounded-md font-medium hover:bg-blue-700">
                    Refresh
                </button>
            </div>

            <div className="overflow-x-auto rounded-lg border border-slate-200 bg-white shadow-sm">
                <table className="min-w-full table-auto border-collapse text-slate-900">
                    <thead>
                        <tr className="bg-slate-50 text-left text-slate-600">
                            <th className="px-4 py-2">ID</th>
                            <th className="px-4 py-2">Category</th>
                            <th className="px-4 py-2">Status</th>
                            <th className="px-4 py-2">Assigned</th>
                            <th className="px-4 py-2">Created</th>
                            <th className="px-4 py-2">Action</th>
                        </tr>
                    </thead>
                    <tbody>
                        {complaints.map(c => (
                            <tr key={c.complaintId} className="border-t border-slate-200 hover:bg-slate-50">
                                <td className="px-4 py-2 font-mono text-sm">{c.complaintId}</td>
                                <td className="px-4 py-2 capitalize">{c.category}</td>
                                <td className="px-4 py-2">
                                    <span className={`px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[c.status] || 'bg-gray-100'}`}>
                                        {c.status}
                                    </span>
                                </td>
                                <td className="px-4 py-2 text-sm">{c.assignedDept || '-'}</td>
                                <td className="px-4 py-2 text-sm">{formatDateValue(c.createdAt)}</td>
                                <td className="px-4 py-2">
                                    <button onClick={() => selectComplaint(c)} className="text-blue-700 hover:text-blue-800 hover:underline text-sm font-medium">
                                        Manage
                                    </button>
                                </td>
                            </tr>
                        ))}
                        {complaints.length === 0 && (
                            <tr><td colSpan={6} className="px-4 py-4 text-slate-500 text-center">No complaints found.</td></tr>
                        )}
                    </tbody>
                </table>
            </div>

            {selected && (
                <div className="border border-slate-200 rounded-lg p-4 bg-white text-slate-900 shadow-sm">
                    <div className="flex items-center justify-between mb-4">
                        <h3 className="text-lg font-semibold">Manage: {selected.complaintId}</h3>
                        <button onClick={() => setSelected(null)} className="text-slate-400 hover:text-slate-700">✕</button>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm text-slate-700 mb-4">
                        <div><strong className="text-slate-900">Status:</strong> {selected.status}</div>
                        <div><strong className="text-slate-900">Category:</strong> {selected.category}</div>
                        <div><strong className="text-slate-900">Assigned:</strong> {selected.assignedDept || 'Unassigned'}</div>
                        <div><strong className="text-slate-900">Created:</strong> {formatDateValue(selected.createdAt)}</div>
                        {selected.closureReason && <div className="col-span-1 md:col-span-2"><strong className="text-slate-900">Closure Reason:</strong> {selected.closureReason}</div>}
                        {selected.rejectedReason && <div className="col-span-1 md:col-span-2"><strong className="text-slate-900">Rejection Reason:</strong> {selected.rejectedReason}</div>}
                    </div>

                    <div className="flex flex-wrap gap-2 mb-4">
                        {selected.status === 'Submitted' && (
                            <>
                                <button disabled={loading} onClick={() => doAction('acknowledge')} className="px-3 py-1.5 bg-blue-600 text-white rounded-md text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed">Acknowledge (ACB)</button>
                                <button disabled={loading} onClick={() => doAction('reject', 'Invalid submission')} className="px-3 py-1.5 bg-red-600 text-white rounded-md text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed">Reject (ACB)</button>
                            </>
                        )}
                        {selected.status === 'Acknowledged' && (
                            <>
                                <button disabled={loading} onClick={() => doAction('assign', 'DeptMSP')} className="px-3 py-1.5 bg-indigo-600 text-white rounded-md text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed">Assign to Dept</button>
                                <button disabled={loading} onClick={() => doAction('updateStatus', 'UnderInvestigation')} className="px-3 py-1.5 bg-yellow-600 text-white rounded-md text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed">Start Investigation</button>
                            </>
                        )}
                        {selected.status === 'UnderInvestigation' && (
                            <>
                                <button disabled={loading} onClick={() => doAction('addEvidence', 'evidence-doc-001')} className="px-3 py-1.5 bg-slate-600 text-white rounded-md text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed">Add Evidence</button>
                                <button disabled={loading} onClick={() => doAction('updateStatus', 'ActionTaken')} className="px-3 py-1.5 bg-purple-600 text-white rounded-md text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed">Mark Action Taken</button>
                            </>
                        )}
                        {selected.status === 'ActionTaken' && (
                            <button disabled={loading} onClick={() => doAction('requestClosure', 'Investigation complete. Recommended closure.')} className="px-3 py-1.5 bg-orange-600 text-white rounded-md text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed">Request Closure (ACB)</button>
                        )}
                        {selected.status === 'PendingClosure' && (
                            <button disabled={loading} onClick={() => doAction('approveClosure')} className="px-3 py-1.5 bg-green-600 text-white rounded-md text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed">Approve Closure (Oversight)</button>
                        )}
                    </div>

                    {message && <div className="text-sm mb-2 text-slate-700">{message}</div>}

                    <div>
                        <h4 className="font-semibold text-sm text-slate-900 mb-2">Audit Trail (Ledger History)</h4>
                        <div className="max-h-48 overflow-y-auto border border-slate-200 rounded p-2 bg-slate-50 text-xs font-mono text-slate-700 space-y-1">
                            {history.length === 0 && <div className="text-slate-400">No history available.</div>}
                            {history.map((h, i) => (
                                <div key={i} className="border-b border-slate-200 pb-1 last:border-b-0">
                                    <div className="text-slate-500">Tx: {h.txId}</div>
                                    <div className="text-slate-700">{formatDateValue(h.timestamp)}</div>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
