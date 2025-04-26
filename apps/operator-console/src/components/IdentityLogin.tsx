'use client';
import { useState, createContext, useContext, ReactNode } from 'react';

interface IdentityContextProps {
    user: string | null;
    login: (userId: string, password: string) => Promise<void>;
}
const IdentityContext = createContext<IdentityContextProps>({ user: null, login: async () => {} });

export function IdentityProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<string | null>(null);
    const login = async (userId: string, password: string) => {
        const res = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ userId, password }),
        });
        if (res.ok) {
            setUser(userId);
        } else {
            throw new Error('Login failed');
        }
    };
    return (
        <IdentityContext.Provider value={{ user, login }}>
            {children}
        </IdentityContext.Provider>
    );
}

export default function IdentityLogin() {
    const { login } = useContext(IdentityContext);
    const [userId, setUserId] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await login(userId, password);
        } catch (err) {
            setError((err as Error).message);
        }
    };

    return (
        <form onSubmit={handleSubmit} className="max-w-sm">
            <label className="block">
                <span>User ID</span>
                <input
                    className="mt-1 block w-full border rounded px-2 py-1"
                    value={userId}
                    onChange={e => setUserId(e.target.value)}
                />
            </label>
            <label className="block mt-4">
                <span>Password</span>
                <input
                    type="password"
                    className="mt-1 block w-full border rounded px-2 py-1"
                    value={password}
                    onChange={e => setPassword(e.target.value)}
                />
            </label>
            {error && <p className="text-red-600 mt-2">{error}</p>}
            <button
                type="submit"
                className="mt-4 px-4 py-2 bg-blue-600 text-white rounded"
            >
                Login
            </button>
        </form>
    );
}