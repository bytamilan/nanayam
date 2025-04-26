'use client';
import { useEffect, useState } from 'react';
import { listChannels } from '@/actions/listChannels';

export default function ChannelTable() {
    const [channels, setChannels] = useState<string[]>([]);
    useEffect(() => { listChannels().then(setChannels); }, []);

    return (
        <table className="min-w-full table-auto">
            <thead><tr><th>Channel</th></tr></thead>
            <tbody>
            {channels.map(c => <tr key={c}><td>{c}</td></tr>)}
            </tbody>
        </table>
    );
}