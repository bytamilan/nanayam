'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';

const ORGS = [
    { value: 'ACBMSP', label: 'ACB (Anti-Corruption Bureau)' },
    { value: 'DeptMSP', label: 'Department' },
    { value: 'OversightMSP', label: 'Oversight' },
    { value: 'JudiciaryMSP', label: 'Judiciary' },
];

export default function SignupForm() {
    const router = useRouter();
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const [org, setOrg] = useState('ACBMSP');
    const [error, setError] = useState<string | null>(null);
    const [submitting, setSubmitting] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        if (password !== confirm) {
            setError('Passwords do not match');
            return;
        }
        setSubmitting(true);
        try {
            const res = await fetch('/api/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password, org }),
            });
            const data = await res.json();
            if (!res.ok) {
                throw new Error(data.error || 'Registration failed');
            }
            router.push('/login');
        } catch (err: any) {
            setError(err.message || 'Registration failed');
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            <div>
                <label className="block text-sm font-medium text-slate-700">Username</label>
                <input
                    type="text"
                    required
                    className="mt-1 block w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                />
            </div>
            <div>
                <label className="block text-sm font-medium text-slate-700">Password</label>
                <input
                    type="password"
                    required
                    className="mt-1 block w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                />
            </div>
            <div>
                <label className="block text-sm font-medium text-slate-700">Confirm Password</label>
                <input
                    type="password"
                    required
                    className="mt-1 block w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                    value={confirm}
                    onChange={(e) => setConfirm(e.target.value)}
                />
            </div>
            <div>
                <label className="block text-sm font-medium text-slate-700">Organization</label>
                <select
                    className="mt-1 block w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                    value={org}
                    onChange={(e) => setOrg(e.target.value)}
                >
                    {ORGS.map((o) => (
                        <option key={o.value} value={o.value}>{o.label}</option>
                    ))}
                </select>
            </div>
            {error && (
                <p className="text-sm text-red-600 bg-red-50 rounded-md px-3 py-2">{error}</p>
            )}
            <button
                type="submit"
                disabled={submitting}
                className="w-full rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
                {submitting ? 'Creating account…' : 'Create account'}
            </button>
        </form>
    );
}
