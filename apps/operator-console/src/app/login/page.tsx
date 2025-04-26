import IdentityLogin from '@/components/IdentityLogin';
export default function LoginPage() {
    return (
        <div className="p-8">
            <h1 className="text-2xl mb-4">Fabric Identity Login</h1>
            <IdentityLogin />
        </div>
    );
}