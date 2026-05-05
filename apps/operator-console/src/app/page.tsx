import AssetTable from '@/components/AssetTable';
import ChannelTable from '@/components/ChannelTable';
import ComplaintTable from '@/components/ComplaintTable';
import ComplaintForm from '@/components/ComplaintForm';

export const dynamic = 'force-dynamic';

export default async function Dashboard() {
    return (
        <div className="p-8 max-w-6xl mx-auto space-y-10">
            <div>
                <h1 className="text-3xl font-semibold mb-2">Anti-Corruption Complaint System</h1>
                <p className="text-slate-400 text-sm">
                    Multi-authority workflow on Hyperledger Fabric. No single department can suppress or alter a complaint.
                </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                <div className="border border-blue-100 rounded-lg p-4 bg-blue-50 text-slate-700 shadow-sm">
                    <strong className="block text-blue-800">ACB</strong>
                    Acknowledge, assign, investigate, request closure
                </div>
                <div className="border border-yellow-100 rounded-lg p-4 bg-yellow-50 text-slate-700 shadow-sm">
                    <strong className="block text-yellow-800">Department</strong>
                    Update status, add evidence (cannot close)
                </div>
                <div className="border border-green-100 rounded-lg p-4 bg-green-50 text-slate-700 shadow-sm">
                    <strong className="block text-green-800">Oversight</strong>
                    Must co-endorse closure. Prevents unilateral suppression.
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="lg:col-span-2">
                    <ComplaintTable />
                </div>
                <div>
                    <ComplaintForm />
                </div>
            </div>

            <div className="border-t border-white/10 pt-6">
                <h2 className="text-xl font-semibold mb-4">Legacy Asset Manager</h2>
                <div className="mb-4">
                    <ChannelTable />
                </div>
                <div>
                    <AssetTable />
                </div>
            </div>
        </div>
    );
}
