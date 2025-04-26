'use client';
import { IdentityProvider } from '@/components/IdentityLogin';
export default function RootLayout({ children }: { children: React.ReactNode }) {
    return (
        <html lang="en">
        <body>
        <IdentityProvider>
            {children}
        </IdentityProvider>
        </body>
        </html>
    );
}