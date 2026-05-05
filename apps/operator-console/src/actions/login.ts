export async function login(userId: string, password: string): Promise<boolean> {
    const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ userId, password }),
    });
    const { success } = await res.json();
    return success;
}