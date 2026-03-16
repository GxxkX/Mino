'use client';

import Image from 'next/image';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { 
  Home, 
  Brain, 
  CheckSquare, 
  Mic, 
  MessageCircle, 
  Settings, 
  Puzzle,
  Users
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useTranslation } from '@/lib/i18n';

export function Sidebar() {
  const pathname = usePathname();
  const t = useTranslation();

  const navItems = [
    { href: '/dashboard', label: t.nav.schedule, icon: Home },
    { href: '/dashboard/memories', label: t.nav.memories, icon: Brain },
    { href: '/dashboard/tasks', label: t.nav.tasks, icon: CheckSquare },
    { href: '/dashboard/audio', label: t.nav.audio, icon: Mic },
    { href: '/dashboard/speakers', label: t.nav.speakers, icon: Users },
    { href: '/dashboard/chat', label: t.nav.chat, icon: MessageCircle },
    { href: '/dashboard/extensions', label: t.nav.extensions, icon: Puzzle },
  ];

  const bottomNavItems = [
    { href: '/dashboard/settings', label: t.nav.settings, icon: Settings },
  ];

  return (
    <aside className="w-56 flex flex-col h-screen sticky top-0 border-r border-border bg-background">
      <div className="px-5 py-6">
        <Link href="/dashboard" className="flex items-center gap-2.5">
          <Image src="/logo.png" alt="Mino" width={28} height={28} className="rounded-md" />
          <span className="text-[15px] font-semibold text-text tracking-tight">Mino</span>
        </Link>
      </div>

      <nav className="flex-1 px-3 space-y-0.5">
        <p className="section-label px-3 pb-2">{t.nav.navigation}</p>
        {navItems.map((item) => {
          const isActive = pathname === item.href || 
            (item.href !== '/dashboard' && pathname.startsWith(item.href));
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center gap-2.5 px-3 py-2 rounded-md text-[13px] transition-colors duration-150 cursor-pointer',
                isActive 
                  ? 'bg-surface-hover text-text font-medium' 
                  : 'text-text-muted hover:text-text-secondary hover:bg-surface'
              )}
            >
              <item.icon className="w-4 h-4" strokeWidth={isActive ? 2 : 1.5} />
              <span>{item.label}</span>
            </Link>
          );
        })}
      </nav>

      <div className="px-3 pb-4 space-y-0.5">
        <div className="border-t border-border mb-3" />
        {bottomNavItems.map((item) => {
          const isActive = pathname.startsWith(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center gap-2.5 px-3 py-2 rounded-md text-[13px] transition-colors duration-150 cursor-pointer',
                isActive 
                  ? 'bg-surface-hover text-text font-medium' 
                  : 'text-text-muted hover:text-text-secondary hover:bg-surface'
              )}
            >
              <item.icon className="w-4 h-4" strokeWidth={isActive ? 2 : 1.5} />
              <span>{item.label}</span>
            </Link>
          );
        })}
      </div>
    </aside>
  );
}
