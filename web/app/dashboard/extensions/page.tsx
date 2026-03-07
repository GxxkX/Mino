'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { useAppStore } from '@/lib/store';
import { useTranslation } from '@/lib/i18n';
import { extensionsApi } from '@/lib/api';
import { ExternalLink, Plus, Trash2, Calendar, MessageSquare, Music, Book, Zap, Globe, Bell, Star } from 'lucide-react';
import type { Extension } from '@/types';

const ICON_MAP: Record<string, React.ElementType> = {
  calendar: Calendar,
  message: MessageSquare,
  music: Music,
  book: Book,
  zap: Zap,
  globe: Globe,
  bell: Bell,
  star: Star,
};

function ExtIcon({ icon }: { icon: string }) {
  const Icon = ICON_MAP[icon.toLowerCase()] ?? Zap;
  return <Icon className="w-4 h-4 text-text-muted" />;
}

export default function ExtensionsPage() {
  const { extensions, setExtensions, updateExtension, removeExtension } = useAppStore();
  const t = useTranslation();
  const [loading, setLoading] = useState(true);
  const [toggling, setToggling] = useState<string | null>(null);
  const [showAdd, setShowAdd] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDesc, setNewDesc] = useState('');
  const [adding, setAdding] = useState(false);

  useEffect(() => {
    extensionsApi.listExtensions()
      .then(setExtensions)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [setExtensions]);

  async function handleToggle(ext: Extension) {
    if (toggling) return;
    setToggling(ext.id);
    try {
      const updated = await extensionsApi.updateExtension(ext.id, { enabled: !ext.enabled });
      updateExtension(ext.id, { enabled: updated.enabled });
    } catch (e) {
      console.error(e);
    } finally {
      setToggling(null);
    }
  }

  async function handleDelete(id: string) {
    try {
      await extensionsApi.deleteExtension(id);
      removeExtension(id);
    } catch (e) {
      console.error(e);
    }
  }

  async function handleAdd() {
    if (!newName.trim()) return;
    setAdding(true);
    try {
      const ext = await extensionsApi.createExtension({ name: newName.trim(), description: newDesc.trim() });
      setExtensions([...extensions, ext]);
      setNewName('');
      setNewDesc('');
      setShowAdd(false);
    } catch (e) {
      console.error(e);
    } finally {
      setAdding(false);
    }
  }

  const enabled = extensions.filter(e => e.enabled);
  const disabled = extensions.filter(e => !e.enabled);

  return (
    <div className="min-h-screen">
      <Header title={t.extensions.title} description={t.extensions.description} />

      <div className="px-8 pb-8 pt-5 space-y-6">
        {loading ? (
          <div className="text-center py-16 text-text-muted text-sm">{t.loading}</div>
        ) : (
          <>
            {extensions.length === 0 && !showAdd ? (
              <div className="text-center py-16 text-text-muted text-sm">{t.extensions.empty}</div>
            ) : (
              <>
                {enabled.length > 0 && (
                  <section>
                    <p className="section-label mb-3">{t.extensions.enabled}</p>
                    <div className="divide-y divide-border-subtle">
                      {enabled.map(ext => (
                        <ExtRow
                          key={ext.id}
                          ext={ext}
                          toggling={toggling === ext.id}
                          onToggle={() => handleToggle(ext)}
                          onDelete={() => handleDelete(ext.id)}
                          configureLabel={t.extensions.configure}
                        />
                      ))}
                    </div>
                  </section>
                )}

                {disabled.length > 0 && (
                  <section>
                    <p className="section-label mb-3">{t.extensions.installed}</p>
                    <div className="divide-y divide-border-subtle">
                      {disabled.map(ext => (
                        <ExtRow
                          key={ext.id}
                          ext={ext}
                          toggling={toggling === ext.id}
                          onToggle={() => handleToggle(ext)}
                          onDelete={() => handleDelete(ext.id)}
                          configureLabel={t.extensions.configure}
                        />
                      ))}
                    </div>
                  </section>
                )}
              </>
            )}

            {showAdd && (
              <div className="border border-border rounded-lg p-4 space-y-3">
                <p className="text-sm font-medium text-text">{t.extensions.addExtension}</p>
                <input
                  className="w-full bg-surface-hover border border-border rounded-md px-3 py-2 text-sm text-text placeholder:text-text-muted outline-none focus:border-cta transition-colors"
                  placeholder={t.extensions.namePlaceholder}
                  value={newName}
                  onChange={e => setNewName(e.target.value)}
                />
                <input
                  className="w-full bg-surface-hover border border-border rounded-md px-3 py-2 text-sm text-text placeholder:text-text-muted outline-none focus:border-cta transition-colors"
                  placeholder={t.extensions.descPlaceholder}
                  value={newDesc}
                  onChange={e => setNewDesc(e.target.value)}
                />
                <div className="flex gap-2">
                  <button
                    onClick={handleAdd}
                    disabled={adding || !newName.trim()}
                    className="px-3 py-1.5 rounded-md bg-cta text-white text-xs font-medium disabled:opacity-50 cursor-pointer hover:opacity-90 transition-opacity"
                  >
                    {adding ? t.extensions.adding : t.extensions.confirmAdd}
                  </button>
                  <button
                    onClick={() => { setShowAdd(false); setNewName(''); setNewDesc(''); }}
                    className="px-3 py-1.5 rounded-md text-xs text-text-muted hover:text-text-secondary transition-colors cursor-pointer"
                  >
                    {t.cancel}
                  </button>
                </div>
              </div>
            )}

            <div className="pt-2 border-t border-border">
              <button
                onClick={() => setShowAdd(v => !v)}
                className="flex items-center gap-2 text-xs text-text-muted hover:text-text-secondary transition-colors cursor-pointer py-2"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>{t.extensions.addExtension}</span>
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

function ExtRow({
  ext,
  toggling,
  onToggle,
  onDelete,
  configureLabel,
}: {
  ext: Extension;
  toggling: boolean;
  onToggle: () => void;
  onDelete: () => void;
  configureLabel: string;
}) {
  return (
    <div className="flex items-center gap-4 py-4 group">
      <div className="w-9 h-9 rounded-md bg-surface-hover flex items-center justify-center flex-shrink-0">
        <ExtIcon icon={ext.icon} />
      </div>

      <div className="flex-1 min-w-0">
        <h3 className="text-sm font-medium text-text">{ext.name}</h3>
        {ext.description && (
          <p className="text-xs text-text-muted mt-0.5">{ext.description}</p>
        )}
      </div>

      <div className="flex items-center gap-3">
        <button className="text-xs text-text-muted hover:text-text-secondary transition-colors cursor-pointer flex items-center gap-1 opacity-0 group-hover:opacity-100">
          <ExternalLink className="w-3 h-3" />
          {configureLabel}
        </button>
        <button
          onClick={onDelete}
          className="text-xs text-text-muted hover:text-red-400 transition-colors cursor-pointer opacity-0 group-hover:opacity-100"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
        <button
          onClick={onToggle}
          disabled={toggling}
          className={`w-9 h-5 rounded-full transition-colors duration-150 cursor-pointer flex items-center disabled:opacity-60 ${
            ext.enabled ? 'bg-cta justify-end' : 'bg-border justify-start'
          }`}
        >
          <div className="w-4 h-4 rounded-full bg-white mx-0.5" />
        </button>
      </div>
    </div>
  );
}
