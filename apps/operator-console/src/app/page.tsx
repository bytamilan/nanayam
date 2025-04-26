import AssetTable from '@/components/AssetTable';
import ChannelTable from '@/components/ChannelTable';

export default async function Dashboard() {
    // you can call server-only helpers here if needed
    return (
        <div className="p-8">
            <h1 className="text-3xl mb-6">Fabric Operator Console</h1>
            <div className="mb-8">
                <ChannelTable />
            </div>
            <div>
                <AssetTable />
            </div>
        </div>
    );
}