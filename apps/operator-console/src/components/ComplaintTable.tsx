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
                <button onClick={load} className="px-3 py-1 bg-blue-600 text-white rounded hover:bg-blue-700">
                    Refresh
                </button>
            </div>

            <div className="overflow-x-auto">
                <table className="min-w-full table-auto border-collapse">
                    <thead>
                        <tr className="bg-gray-100 text-left">
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
                            <tr key={c.complaintId} className="border-t hover:bg-gray-50">
                                <td className="px-4 py-2 font-mono text-sm">{c.complaintId}</td>
                                <td className="px-4 py-2 capitalize">{c.category}</td>
                                <td className="px-4 py-2">
                                    <span className={`px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[c.status] || 'bg-gray-100'}`}>
                                        {c.status}
                                    </span>
                                </td>
                                <td className="px-4 py-2 text-sm">{c.assignedDept || '-'}</td>
                                <td className="px-4 py-2 text-sm">{new Date(c.createdAt).toLocaleString()}</td>
                                <td className="px-4 py-2">
                                    <button onClick={() => selectComplaint(c)} className="text-blue-600 hover:underline text-sm">
                                        Manage
                                    </button>
                                </td>
                            </tr>
                        ))}
                        {complaints.length === 0 && (
                            <tr><td colSpan={6} className="px-4 py-4 text-gray-500 text-center">No complaints found.</td></tr>
                        )}
                    </tbody>
                </table>
            </div>

            {selected && (
                <div className="border rounded-lg p-4 bg-white shadow-sm">
                    <div className="flex items-center justify-between mb-4">
                        <h3 className="text-lg font-semibold">Manage: {selected.complaintId}</h3>
                        <button onClick={() => setSelected(null)} className="text-gray-500 hover:text-gray-700">✕</button>
                    </div>

                    <div className="grid grid-cols-2 gap-4 text-sm mb-4">
                        <div><strong>Status:</strong> {selected.status}</div>
                        <div><strong>Category:</strong> {selected.category}</div>
                        <div><strong>Assigned:</strong> {selected.assignedDept || 'Unassigned'}</div>
                        <div><strong>Created:</strong> {new Date(selected.createdAt).toLocaleString()}</div>
                        {selected.closureReason && <div className="col-span-2"><strong>Closure Reason:</strong> {selected.closureReason}</div>}
                        {selected.rejectedReason && <div className="col-span-2"><strong>Rejection Reason:</strong> {selected.rejectedReason}</div>}
                    </div>

                    <div className="flex flex-wrap gap-2 mb-4">
                        {selected.status === 'Submitted' && (
                            <>
                                <button disabled={loading} onClick={() => doAction('acknowledge')} className="px-3 py-1 bg-blue-600 text-white rounded text-sm disabled:opacity-50">Acknowledge (ACB)</button>
                                <button disabled={loading} onClick={() => doAction('reject', 'Invalid submission')} className="px-3 py-1 bg-red-600 text-white rounded text-sm disabled:opacity-50">Reject (ACB)</button>
                            </>
                        )}
                        {selected.status === 'Acknowledged' && (
                            <>
                                <button disabled={loading} onClick={() => doAction('assign', 'DeptMSP')} className="px-3 py-1 bg-indigo-600 text-white rounded text-sm disabled:opacity-50">Assign to Dept</button>
                                <button disabled={loading} onClick={() => doAction('updateStatus', 'UnderInvestigation')} className="px-3 py-1 bg-yellow-600 text-white rounded text-sm disabled:opacity-50">Start Investigation</button>
                            </>
                        )}
                        {selected.status === 'UnderInvestigation' && (
                            <>
                                <button disabled={loading} onClick={() => doAction('addEvidence', 'evidence-doc-001')} className="px-3 py-1 bg-gray-600 text-white rounded text-sm disabled:opacity-50">Add Evidence</button>
                                <button disabled={loading} onClick={() => doAction('updateStatus', 'ActionTaken')} className="px-3 py-1 bg-purple-600 text-white rounded text-sm disabled:opacity-50">Mark Action Taken</button>
                            </>
                        )}
                        {selected.status === 'ActionTaken' && (
                            <button disabled={loading} onClick={() => doAction('requestClosure', 'Investigation complete. Recommended closure.')} className="px-3 py-1 bg-orange-600 text-white rounded text-sm disabled:opacity-50">Request Closure (ACB)</button>
                        )}
                        {selected.status === 'PendingClosure' && (
                            <button disabled={loading} onClick={() => doAction('approveClosure')} className="px-3 py-1 bg-green-600 text-white rounded text-sm disabled:opacity-50">Approve Closure (Oversight)</button>
                        )}
                    </div>

                    {message && <div className="text-sm mb-2 text-gray-700">{message}</div>}

                    <div>
                        <h4 className="font-semibold text-sm mb-2">Audit Trail (Ledger History)</h4>
                        <div className="max-h-48 overflow-y-auto border rounded p-2 bg-gray-50 text-xs font-mono space-y-1">
                            {history.length === 0 && <div className="text-gray-400">No history available.</div>}
                            {history.map((h, i) => (
                                <div key={i} className="border-b pb-1">
                                    <div className="text-gray-500">Tx: {h.txId}</div>
                                    <div className="text-gray-700">{h.timestamp}</div>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
