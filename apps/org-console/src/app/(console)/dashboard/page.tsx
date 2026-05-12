import { redirect } from 'next/navigation';

export const dynamic = 'force-dynamic';

export default function DashboardPage() {
    // For now, redirect to complaints since that's where the main action is
    // In a full implementation this would show stat cards + recent activity
    redirect('/complaints');
}
