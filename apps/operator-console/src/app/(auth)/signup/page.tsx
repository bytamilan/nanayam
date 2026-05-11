import Link from 'next/link';
import SignupForm from '@/components/auth/SignupForm';

export const dynamic = 'force-dynamic';

export default function SignupPage() {
    return (
        <div className="flex min-h-screen items-center justify-center bg-slate-50 p-4">
            <div className="w-full max-w-sm rounded-xl bg-white p-8 shadow-lg">
                <h1 className="text-2xl font-semibold text-slate-900">Create Account</h1>
                <p className="mt-1 text-sm text-slate-500">Register for your organization console</p>
                <div className="mt-6">
                    <SignupForm />
                </div>
                <p className="mt-4 text-center text-sm text-slate-500">
                    Already have an account?{' '}
                    <Link href="/login" className="text-blue-600 hover:underline">
                        Sign in
                    </Link>
                </p>
            </div>
        </div>
    );
}
