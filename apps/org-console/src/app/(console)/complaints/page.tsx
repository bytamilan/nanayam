import ComplaintTable from '@/components/ComplaintTable';
import ComplaintForm from '@/components/ComplaintForm';

export const dynamic = 'force-dynamic';

export default function ComplaintsPage() {
    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-2xl font-semibold text-slate-900">Complaints</h1>
                <p className="text-sm text-slate-500 mt-1">
                    Manage and track anti-corruption complaints across the network.
                </p>
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="lg:col-span-2">
                    <ComplaintTable />
                </div>
                <div>
                    <ComplaintForm />
                </div>
            </div>
        </div>
    );
}
