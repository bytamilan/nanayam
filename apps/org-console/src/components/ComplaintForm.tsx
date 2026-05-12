'use client';
import { useState } from 'react';
import { submitComplaint } from '../lib/api';

export default function ComplaintForm() {
    const [complaintId, setComplaintId] = useState('');
    const [category, setCategory] = useState('bribery');
    const [citizenHash, setCitizenHash] = useState('');
    const [descriptionHash, setDescriptionHash] = useState('');
    const [attachmentsRef, setAttachmentsRef] = useState('');
    const [loading, setLoading] = useState(false);
    const [message, setMessage] = useState('');
    const messageTone = message.includes('successfully') ? 'text-green-700' : 'text-red-600';

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        setLoading(true);
        setMessage('');
        const ok = await submitComplaint({
            complaintId: complaintId || `COMP-${Date.now()}`,
            category,
            citizenHash: citizenHash || 'sha256:anonymous',
            descriptionHash: descriptionHash || 'sha256:description',
            attachmentsRef: attachmentsRef || 'ipfs:QmExample',
        });
        setLoading(false);
        if (ok) {
            setMessage('Complaint submitted successfully!');
            setComplaintId('');
            setCitizenHash('');
            setDescriptionHash('');
            setAttachmentsRef('');
        } else {
            setMessage('Submission failed. Check gateway logs.');
        }
    }

    return (
        <div className="border border-slate-200 rounded-lg p-4 bg-white text-slate-900 shadow-sm">
            <h3 className="text-lg font-semibold mb-4">Submit New Complaint</h3>
            <form onSubmit={handleSubmit} className="space-y-3">
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">Complaint ID</label>
                    <input type="text" value={complaintId} onChange={e => setComplaintId(e.target.value)}
                        placeholder="Leave blank for auto-generated"
                        className="w-full border border-slate-300 rounded px-3 py-2 text-sm bg-white text-slate-900 placeholder:text-slate-400 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-200" />
                </div>
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">Category</label>
                    <select value={category} onChange={e => setCategory(e.target.value)}
                        className="w-full border border-slate-300 rounded px-3 py-2 text-sm bg-white text-slate-900 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-200">
                        <option value="bribery">Bribery</option>
                        <option value="abuse">Abuse of Power</option>
                        <option value="delay">Unreasonable Delay</option>
                    </select>
                </div>
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">Citizen Hash (PII)</label>
                    <input type="text" value={citizenHash} onChange={e => setCitizenHash(e.target.value)}
                        placeholder="sha256:..."
                        className="w-full border border-slate-300 rounded px-3 py-2 text-sm bg-white text-slate-900 placeholder:text-slate-400 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-200" />
                </div>
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">Description Hash</label>
                    <input type="text" value={descriptionHash} onChange={e => setDescriptionHash(e.target.value)}
                        placeholder="sha256:..."
                        className="w-full border border-slate-300 rounded px-3 py-2 text-sm bg-white text-slate-900 placeholder:text-slate-400 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-200" />
                </div>
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">Attachments Ref (IPFS/S3)</label>
                    <input type="text" value={attachmentsRef} onChange={e => setAttachmentsRef(e.target.value)}
                        placeholder="ipfs:Qm..."
                        className="w-full border border-slate-300 rounded px-3 py-2 text-sm bg-white text-slate-900 placeholder:text-slate-400 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-200" />
                </div>
                <button type="submit" disabled={loading}
                    className="px-4 py-2 bg-green-600 text-white rounded-md font-medium hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed">
                    {loading ? 'Submitting...' : 'Submit Complaint'}
                </button>
                {message && <div className={`text-sm font-medium ${messageTone}`}>{message}</div>}
            </form>
        </div>
    );
}
