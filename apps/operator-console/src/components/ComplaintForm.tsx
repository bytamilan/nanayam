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
        <div className="border rounded-lg p-4 bg-white shadow-sm">
            <h3 className="text-lg font-semibold mb-4">Submit New Complaint</h3>
            <form onSubmit={handleSubmit} className="space-y-3">
                <div>
                    <label className="block text-sm font-medium">Complaint ID</label>
                    <input type="text" value={complaintId} onChange={e => setComplaintId(e.target.value)}
                        placeholder="Leave blank for auto-generated"
                        className="w-full border rounded px-3 py-2 text-sm" />
                </div>
                <div>
                    <label className="block text-sm font-medium">Category</label>
                    <select value={category} onChange={e => setCategory(e.target.value)}
                        className="w-full border rounded px-3 py-2 text-sm">
                        <option value="bribery">Bribery</option>
                        <option value="abuse">Abuse of Power</option>
                        <option value="delay">Unreasonable Delay</option>
                    </select>
                </div>
                <div>
                    <label className="block text-sm font-medium">Citizen Hash (PII)</label>
                    <input type="text" value={citizenHash} onChange={e => setCitizenHash(e.target.value)}
                        placeholder="sha256:..."
                        className="w-full border rounded px-3 py-2 text-sm" />
                </div>
                <div>
                    <label className="block text-sm font-medium">Description Hash</label>
                    <input type="text" value={descriptionHash} onChange={e => setDescriptionHash(e.target.value)}
                        placeholder="sha256:..."
                        className="w-full border rounded px-3 py-2 text-sm" />
                </div>
                <div>
                    <label className="block text-sm font-medium">Attachments Ref (IPFS/S3)</label>
                    <input type="text" value={attachmentsRef} onChange={e => setAttachmentsRef(e.target.value)}
                        placeholder="ipfs:Qm..."
                        className="w-full border rounded px-3 py-2 text-sm" />
                </div>
                <button type="submit" disabled={loading}
                    className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50">
                    {loading ? 'Submitting...' : 'Submit Complaint'}
                </button>
                {message && <div className="text-sm text-gray-700">{message}</div>}
            </form>
        </div>
    );
}
