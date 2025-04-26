'use client';
import { useEffect, useState } from 'react';
import { listAssets } from '../actions/listAssets';

export default function AssetTable() {
    const [assets, setAssets] = useState<string[]>([]);
    useEffect(() => { listAssets().then(setAssets); }, []);

    return (
        <table className="min-w-full table-auto">
            <thead><tr><th>Asset ID</th></tr></thead>
            <tbody>
            {assets.map(id => <tr key={id}><td>{id}</td></tr>)}
            </tbody>
        </table>
    );
}