'use client';

import { useRouter } from 'next/navigation';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { useAppStore } from '@/lib/store';
import { useTranslation } from '@/lib/i18n';
import { signOut } from '@/lib/api/auth';
import { updateSTTConfig } from '@/lib/api/settings';
import { Cloud, Key, Shield, Globe, ChevronRight, Mic, LogOut, Brain, Plug, Languages } from 'lucide-react';
import Link from 'next/link';

export default function SettingsPage() {
  const { settings, setSettings, user, setUser } = useAppStore();
  const router = useRouter();
  const t = useTranslation();

  const gain = settings.recordingGain ?? 1.0;

  function toggleLanguage() {
    setSettings({ language: settings.language === 'zh-CN' ? 'en' : 'zh-CN' });
  }

  async function handleSignOut() {
    try {
      await signOut();
    } catch {
      // clear tokens even if API fails
    }
    setUser(null);
    router.push('/auth/login');
  }

  const sections = [
    {
      title: t.settings.general,
      items: [
        {
          icon: Globe,
          label: t.settings.language,
          value: settings.language === 'zh-CN' ? '中文' : 'English',
          onClick: toggleLanguage,
        },
      ],
    },
    {
      title: t.settings.cloudSync,
      items: [
        {
          icon: Cloud,
          label: t.settings.cloudSettings,
          value: settings.cloudSync ? t.settings.cloudEnabled : t.settings.cloudDisabled,
          href: '/dashboard/settings/cloud',
        },
      ],
    },
    {
      title: t.settings.config,
      items: [
        {
          icon: Brain,
          label: t.settings.llmProvider,
          value: settings.llmProvider.toUpperCase(),
          href: '/dashboard/settings/llm',
        },
        {
          icon: Plug,
          label: t.mcpSettings.title,
          value: settings.mcpEnabled ? t.settings.cloudEnabled : t.settings.cloudDisabled,
          href: '/dashboard/settings/mcp',
        },
      ],
    },
    {
      title: t.settings.security,
      items: [
        {
          icon: Shield,
          label: t.settings.changePassword,
          value: '',
          href: '/dashboard/settings/security',
        },
      ],
    },
  ];

  return (
    <div className="min-h-screen">
      <Header title={t.settings.title} />

      <div className="px-8 pb-8 max-w-xl space-y-8">
        {/* Profile section */}
        <div className="flex items-center gap-4 py-4 border-b border-border">
          <div className="w-11 h-11 rounded-full bg-surface-hover flex items-center justify-center">
            <span className="text-sm font-semibold text-cta">
              {user?.displayName?.charAt(0).toUpperCase() ?? 'U'}
            </span>
          </div>
          <div>
            <h3 className="text-sm font-medium text-text">{user?.displayName ?? t.settings.user}</h3>
            <p className="text-xs text-text-muted">@{user?.username ?? 'unknown'}</p>
          </div>
        </div>

        {/* Recording section */}
        <div>
          <p className="section-label mb-3">{t.settings.recording}</p>
          <div className="py-3">
            <div className="flex items-center gap-3 mb-3">
              <Mic className="w-4 h-4 text-text-muted flex-shrink-0" />
              <span className="text-sm text-text">{t.settings.recordingGain}</span>
              <span className="ml-auto text-xs text-text-muted tabular-nums w-12 text-right">
                {gain === 1.0 ? t.settings.gainStandard : `${Math.round(gain * 100)}%`}
              </span>
            </div>
            <div className="flex items-center gap-3 pl-7">
              <span className="text-[11px] text-text-muted">{t.settings.gainMute}</span>
              <input
                type="range"
                min="0"
                max="3"
                step="0.1"
                value={gain}
                onChange={(e) => setSettings({ recordingGain: parseFloat(e.target.value) })}
                className="flex-1 h-1 appearance-none bg-border rounded-full outline-none cursor-pointer
                  [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-3.5 [&::-webkit-slider-thumb]:h-3.5
                  [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-cta [&::-webkit-slider-thumb]:shadow-[0_0_6px_rgba(163,230,53,0.4)]
                  [&::-webkit-slider-thumb]:cursor-pointer
                  [&::-moz-range-thumb]:w-3.5 [&::-moz-range-thumb]:h-3.5 [&::-moz-range-thumb]:rounded-full
                  [&::-moz-range-thumb]:bg-cta [&::-moz-range-thumb]:border-0 [&::-moz-range-thumb]:cursor-pointer"
              />
              <span className="text-[11px] text-text-muted">3x</span>
            </div>
            <p className="text-[11px] text-text-muted mt-2 pl-7">
              {t.settings.gainWarning}
            </p>
          </div>
          <div className="flex items-center gap-3 py-3 border-t border-border-subtle">
            <Languages className="w-4 h-4 text-text-muted flex-shrink-0" />
            <span className="text-sm text-text">{t.settings.whisperLanguage}</span>
            <select
              value={settings.whisperLanguage ?? ''}
              onChange={(e) => {
                const lang = e.target.value;
                setSettings({ whisperLanguage: lang });
                updateSTTConfig({ whisper_language: lang }).catch(console.error);
              }}
              className="ml-auto text-xs bg-transparent text-text-muted border border-border rounded-lg px-3 py-1.5
                cursor-pointer outline-none transition-colors duration-200
                focus:border-text-secondary focus:shadow-[0_0_0_3px_rgba(15,23,42,0.12)]
                appearance-none"
              aria-label={t.settings.whisperLanguage}
            >
              <option value="">{t.settings.whisperLanguageAuto}</option>
              <option value="zh">中文</option>
              <option value="en">English</option>
              <option value="ja">日本語</option>
              <option value="ko">한국어</option>
              <option value="es">Español</option>
              <option value="fr">Français</option>
              <option value="de">Deutsch</option>
              <option value="ru">Русский</option>
              <option value="pt">Português</option>
              <option value="ar">العربية</option>
            </select>
          </div>
        </div>

        {sections.map((section) => (
          <div key={section.title}>
            <p className="section-label mb-3">{section.title}</p>
            <div className="divide-y divide-border-subtle">
              {section.items.map((item) => {
                const content = (
                  <div className="flex items-center justify-between py-3 group cursor-pointer">
                    <div className="flex items-center gap-3">
                      <item.icon className="w-4 h-4 text-text-muted" />
                      <span className="text-sm text-text">{item.label}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      {item.value && (
                        <span className="text-xs text-text-muted">{item.value}</span>
                      )}
                      <ChevronRight className="w-3.5 h-3.5 text-text-muted/50 group-hover:text-text-muted transition-colors" />
                    </div>
                  </div>
                );

                if ('href' in item && item.href) {
                  return (
                    <Link key={item.label} href={item.href}>
                      {content}
                    </Link>
                  );
                }

                return (
                  <div
                    key={item.label}
                    onClick={'onClick' in item ? item.onClick : undefined}
                  >
                    {content}
                  </div>
                );
              })}
            </div>
          </div>
        ))}

        <div className="pt-4">
          <Button
            variant="ghost"
            className="w-full text-accent-rose hover:text-accent-rose"
            onClick={handleSignOut}
          >
            <LogOut className="w-3.5 h-3.5" />
            {t.settings.signOut}
          </Button>
        </div>
      </div>
    </div>
  );
}
