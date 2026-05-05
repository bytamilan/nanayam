export interface Asset { assetId: string; color: string; size: number; }

export async function listAssets(): Promise<string[]> {
    const res = await fetch('/api/ListAssets');
    const json = await res.json();
    return json.assetIds;
}

export async function createAsset(asset: Asset): Promise<boolean> {
    const res = await fetch('/api/CreateAsset', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(asset)
    });
    const json = await res.json();
    return json.success;
}

// ---------------------------------------------------------------------------
// Complaint APIs
// ---------------------------------------------------------------------------

export interface Complaint {
    complaintId: string;
    category: string;
    status: string;
    assignedDept: string;
    attachmentsRef: string[];
    createdAt: string;
    updatedAt: string;
    resolvedAt?: string;
    rejectedReason?: string;
    closureReason?: string;
    requestedBy?: string;
}

export async function listComplaints(): Promise<Complaint[]> {
    const res = await fetch('/api/complaints');
    const json = await res.json();
    return json.complaints || [];
}

export async function getComplaint(id: string): Promise<Complaint | null> {
    const res = await fetch(`/api/complaints/${id}`);
    if (!res.ok) return null;
    return res.json();
}

export async function submitComplaint(data: {
    complaintId: string;
    category: string;
    citizenHash: string;
    descriptionHash: string;
    attachmentsRef: string;
}): Promise<boolean> {
    const res = await fetch('/api/complaints/submit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    });
    const json = await res.json();
    return json.success;
}

export async function updateComplaint(data: {
    complaintId: string;
    action: string;
    value?: string;
}): Promise<boolean> {
    const res = await fetch('/api/complaints/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    });
    const json = await res.json();
    return json.success;
}

export async function getComplaintHistory(id: string): Promise<any[]> {
    const res = await fetch(`/api/complaints/${id}/history`);
    if (!res.ok) return [];
    const json = await res.json();
    try {
        return JSON.parse(json.data || '[]');
    } catch {
        return [];
    }
}
