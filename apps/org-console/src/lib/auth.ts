export interface AuthUser {
    id: string;
    username: string;
    org: string;
    role: string;
    createdAt: string;
}

export function getCookie(name: string): string | null {
    if (typeof document === 'undefined') return null;
    const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
    return match ? decodeURIComponent(match[2]) : null;
}

export function deleteCookie(name: string) {
    if (typeof document === 'undefined') return;
    document.cookie = name + '=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
}

export function orgColorClass(org: string): string {
    switch (org) {
        case 'ACBMSP': return 'text-blue-600';
        case 'DeptMSP': return 'text-yellow-600';
        case 'OversightMSP': return 'text-green-600';
        case 'JudiciaryMSP': return 'text-purple-600';
        default: return 'text-slate-600';
    }
}

export function orgBgClass(org: string): string {
    switch (org) {
        case 'ACBMSP': return 'bg-blue-600';
        case 'DeptMSP': return 'bg-yellow-600';
        case 'OversightMSP': return 'bg-green-600';
        case 'JudiciaryMSP': return 'bg-purple-600';
        default: return 'bg-slate-600';
    }
}

export function orgBorderClass(org: string): string {
    switch (org) {
        case 'ACBMSP': return 'border-blue-600';
        case 'DeptMSP': return 'border-yellow-600';
        case 'OversightMSP': return 'border-green-600';
        case 'JudiciaryMSP': return 'border-purple-600';
        default: return 'border-slate-600';
    }
}
