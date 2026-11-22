import { Button } from '@/components/ui/button';

export default function Home() {
    return (
        <div className="flex min-h-screen flex-col items-center justify-center p-24">
            <h1 className="text-3xl font-bold mb-4">Email Platform</h1>
            <p className="mb-4">Welcome to the multi-tenant email automation platform</p>
            <Button variant="outline">Get Started</Button>
        </div>
    );
}
