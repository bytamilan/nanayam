'use client';
import { useEffect, useState } from 'react';
import { listAssets } from '../actions/listAssets';

export default function AssetTable() {
    const [assets, setAssets] = useState<string[]>([]);
    useEffect(() => { listAssets().then(setAssets); }, []);

    return (
        <div className="overflow-hidden rounded-lg border border-slate-200 bg-white text-slate-900 shadow-sm">
            <table className="min-w-full table-auto border-collapse">
                <thead>
                    <tr className="bg-slate-50 text-left text-slate-600">
                        <th className="px-4 py-2">Asset ID</th>
                    </tr>
                </thead>
                <tbody>
                    {assets.map(id => (
                        <tr key={id} className="border-t border-slate-200">
                            <td className="px-4 py-2 font-mono text-sm">{id}</td>
                        </tr>
                    ))}
                    {assets.length === 0 && (
                        <tr>
                            <td className="px-4 py-3 text-slate-500">No assets found.</td>
                        </tr>
                    )}
                </tbody>
            </table>
        </div>
    );
}