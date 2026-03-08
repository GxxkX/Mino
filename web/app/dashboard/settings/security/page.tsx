'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useTranslation } from '@/lib/i18n';
import { changePassword, changeUsername, signOut } from '@/lib/api/auth';
import { Shield, User, Check, AlertCircle } from 'lucide-react';

export default function SecuritySettingsPage() {
  const t = useTranslation();
  const router = useRouter();

  // --- Username form ---
  const [newUsername, setNewUsername] = useState('');
  const [usernamePassword, setUsernamePassword] = useState('');
  const [usernameLoading, setUsernameLoading] = useState(false);
  const [usernameMessage, setUsernameMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const canSubmitUsername =
    newUsername.length >= 2 &&
    usernamePassword.length > 0 &&
    !usernameLoading;

  async function handleUsernameSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmitUsername) return;

    setUsernameLoading(true);
    setUsernameMessage(null);

    try {
      await changeUsername(newUsername, usernamePassword);
      setUsernameMessage({ type: 'success', text: t.securitySettings.usernameSuccess });
      // Sign out after a short delay so user sees the success message
      setTimeout(async () => {
        await signOut();
        router.push('/auth/login');
      }, 1500);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t.securitySettings.usernameError;
      setUsernameMessage({ type: 'error', text: msg });
    } finally {
      setUsernameLoading(false);
    }
  }

  // --- Password form ---
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const canSubmit =
    oldPassword.length > 0 &&
    newPassword.length >= 6 &&
    newPassword === confirmPassword &&
    !loading;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;

    setLoading(true);
    setMessage(null);

    try {
      await changePassword(oldPassword, newPassword);
      setMessage({ type: 'success', text: t.securitySettings.success });
      setOldPassword('');
      setNewPassword('');
      setConfirmPassword('');
      // Sign out after a short delay so user sees the success message
      setTimeout(async () => {
        await signOut();
        router.push('/auth/login');
      }, 1500);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : t.securitySettings.error;
      setMessage({ type: 'error', text: msg });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen">
      <Header title={t.securitySettings.title} backHref="/dashboard/settings" />

      <div className="px-8 pb-8 max-w-xl space-y-10">
        {/* ── Change Username ── */}
        <section>
          <div className="flex items-center gap-3 py-4 border-b border-border">
            <User className="w-4 h-4 text-text-muted" />
            <div>
              <h3 className="text-sm font-medium text-text">{t.securitySettings.changeUsernameTitle}</h3>
              <p className="text-xs text-text-muted mt-0.5">{t.securitySettings.changeUsernameDesc}</p>
            </div>
          </div>

          <form onSubmit={handleUsernameSubmit} className="space-y-4 mt-4">
            <Input
              label={t.securitySettings.newUsername}
              type="text"
              value={newUsername}
              onChange={(e) => setNewUsername(e.target.value)}
              placeholder={t.securitySettings.newUsernamePlaceholder}
              autoComplete="username"
            />
            {newUsername.length > 0 && newUsername.length < 2 && (
              <p className="text-[11px] text-accent-rose -mt-2">{t.securitySettings.usernameTooShort}</p>
            )}
            <Input
              label={t.securitySettings.usernamePassword}
              type="password"
              value={usernamePassword}
              onChange={(e) => setUsernamePassword(e.target.value)}
              placeholder={t.securitySettings.usernamePasswordPlaceholder}
              autoComplete="current-password"
            />

            {usernameMessage && (
              <div
                className={`flex items-center gap-2 text-xs px-3 py-2 rounded-md transition-colors duration-200 ${
                  usernameMessage.type === 'success'
                    ? 'bg-cta/10 text-cta'
                    : 'bg-accent-rose/10 text-accent-rose'
                }`}
              >
                {usernameMessage.type === 'success' ? (
                  <Check className="w-3.5 h-3.5" />
                ) : (
                  <AlertCircle className="w-3.5 h-3.5" />
                )}
                {usernameMessage.text}
              </div>
            )}

            <Button type="submit" disabled={!canSubmitUsername} className="w-full cursor-pointer">
              {usernameLoading ? t.securitySettings.changingUsername : t.securitySettings.submitUsername}
            </Button>
          </form>
        </section>

        {/* ── Change Password ── */}
        <section>
          <div className="flex items-center gap-3 py-4 border-b border-border">
            <Shield className="w-4 h-4 text-text-muted" />
            <div>
              <h3 className="text-sm font-medium text-text">{t.securitySettings.changePasswordTitle}</h3>
              <p className="text-xs text-text-muted mt-0.5">{t.securitySettings.changePasswordDesc}</p>
            </div>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4 mt-4">
            <Input
              label={t.securitySettings.currentPassword}
              type="password"
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              placeholder={t.securitySettings.currentPasswordPlaceholder}
              autoComplete="current-password"
            />
            <Input
              label={t.securitySettings.newPassword}
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder={t.securitySettings.newPasswordPlaceholder}
              autoComplete="new-password"
            />
            {newPassword.length > 0 && newPassword.length < 6 && (
              <p className="text-[11px] text-accent-rose -mt-2">{t.securitySettings.passwordTooShort}</p>
            )}
            <Input
              label={t.securitySettings.confirmPassword}
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder={t.securitySettings.confirmPasswordPlaceholder}
              autoComplete="new-password"
            />
            {confirmPassword.length > 0 && newPassword !== confirmPassword && (
              <p className="text-[11px] text-accent-rose -mt-2">{t.securitySettings.passwordMismatch}</p>
            )}

            {message && (
              <div
                className={`flex items-center gap-2 text-xs px-3 py-2 rounded-md transition-colors duration-200 ${
                  message.type === 'success'
                    ? 'bg-cta/10 text-cta'
                    : 'bg-accent-rose/10 text-accent-rose'
                }`}
              >
                {message.type === 'success' ? (
                  <Check className="w-3.5 h-3.5" />
                ) : (
                  <AlertCircle className="w-3.5 h-3.5" />
                )}
                {message.text}
              </div>
            )}

            <Button type="submit" disabled={!canSubmit} className="w-full cursor-pointer">
              {loading ? t.securitySettings.changing : t.securitySettings.submit}
            </Button>
          </form>
        </section>
      </div>
    </div>
  );
}
