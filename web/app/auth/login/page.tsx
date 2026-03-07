'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { signIn } from '@/lib/api/auth';
import { useTranslation } from '@/lib/i18n';

export default function LoginPage() {
  const router = useRouter();
  const t = useTranslation();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await signIn(username, password);
      router.push('/dashboard');
    } catch (err) {
      setError(err instanceof Error ? err.message : t.auth.loginError);
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center px-4">
      <div className="w-full max-w-sm">
        <div className="flex items-center gap-2.5 mb-10">
          <Image src="/logo.png" alt="Mino" width={32} height={32} className="rounded-lg" />
          <span className="text-base font-semibold text-text tracking-tight">Mino</span>
        </div>

        <h1 className="text-xl font-semibold text-text mb-1">{t.auth.welcome}</h1>
        <p className="text-sm text-text-muted mb-8">{t.auth.subtitle}</p>

        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label={t.auth.username}
            placeholder={t.auth.usernamePlaceholder}
            value={username}
            onChange={e => setUsername(e.target.value)}
            autoComplete="username"
            autoFocus
          />
          <Input
            label={t.auth.password}
            type="password"
            placeholder={t.auth.passwordPlaceholder}
            value={password}
            onChange={e => setPassword(e.target.value)}
            autoComplete="current-password"
          />

          {error && (
            <p className="text-xs text-accent-rose">{error}</p>
          )}

          <Button
            type="submit"
            size="lg"
            className="w-full mt-2"
            disabled={loading || !username || !password}
          >
            {loading ? t.auth.loggingIn : t.auth.loginButton}
          </Button>
        </form>
      </div>
    </div>
  );
}
