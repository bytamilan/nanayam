'use client';
import { useEffect, useState } from 'react';
import { listChannels } from '@/actions/listChannels';

export default function ChannelTable() {
    const [channels, setChannels] = useState<string[]>([]);
    useEffect(() => { listChannels().then(setChannels); }, []);

    return (
        <div className="overflow-hidden rounded-lg border border-slate-200 bg-white text-slate-900 shadow-sm">
            <table className="min-w-full table-auto border-collapse">
                <thead>
                    <tr className="bg-slate-50 text-left text-slate-600">
                        <th className="px-4 py-2">Channel</th>
                    </tr>
                </thead>
                <tbody>
                    {channels.map(c => (
                        <tr key={c} className="border-t border-slate-200">
                            <td className="px-4 py-2">{c}</td>
                        </tr>
                    ))}
                    {channels.length === 0 && (
                        <tr>
                            <td className="px-4 py-3 text-slate-500">No channels found.</td>
                        </tr>
                    )}
                </tbody>
            </table>
        </div>
    );
}