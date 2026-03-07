'use client';

import { useState } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useAppStore } from '@/lib/store';
import { useTranslation } from '@/lib/i18n';
import { Plus, Trash2, Server, Plug } from 'lucide-react';

interface MCPConfig {
  id: string;
  name: string;
  type: string;
  endpoint: string;
  enabled: boolean;
}

export default function MCPSettingsPage() {
  const { settings, setSettings } = useAppStore();
  const t = useTranslation();
  const [configs, setConfigs] = useState<MCPConfig[]>([
    {
      id: '1',
      name: 'Filesystem',
      type: 'filesystem',
      endpoint: '/tmp/mino-files',
      enabled: true,
    },
  ]);

  const addConfig = () => {
    setConfigs([
      ...configs,
      {
        id: Date.now().toString(),
        name: '',
        type: '',
        endpoint: '',
        enabled: true,
      },
    ]);
  };

  const removeConfig = (id: string) => {
    setConfigs(configs.filter(c => c.id !== id));
  };

  const updateConfig = (id: string, field: keyof MCPConfig, value: string | boolean) => {
    setConfigs(configs.map(c => 
      c.id === id ? { ...c, [field]: value } : c
    ));
  };

  return (
    <div className="min-h-screen">
      <Header title={t.mcpSettings.title} backHref="/dashboard/settings" />
      
      <div className="px-8 pb-8 max-w-xl space-y-8">
        <div className="flex items-center justify-between py-4 border-b border-border">
          <div className="flex items-center gap-3">
            <Plug className="w-4 h-4 text-text-muted" />
            <div>
              <h3 className="text-sm font-medium text-text">{t.mcpSettings.protocol}</h3>
              <p className="text-xs text-text-muted mt-0.5">{t.mcpSettings.protocolDesc}</p>
            </div>
          </div>
          <button
            onClick={() => setSettings({ mcpEnabled: !settings.mcpEnabled })}
            className={`w-9 h-5 rounded-full transition-colors duration-150 cursor-pointer flex items-center ${
              settings.mcpEnabled ? 'bg-cta justify-end' : 'bg-border justify-start'
            }`}
          >
            <div className="w-4 h-4 rounded-full bg-white mx-0.5" />
          </button>
        </div>

        <div className="flex items-center justify-between">
          <p className="section-label">{t.mcpSettings.servers}</p>
          <Button size="sm" onClick={addConfig}>
            <Plus className="w-3.5 h-3.5" />
            {t.mcpSettings.add}
          </Button>
        </div>

        <div className="space-y-6">
          {configs.map((config) => (
            <div key={config.id} className="border border-border rounded-lg p-4 space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Server className="w-3.5 h-3.5 text-text-muted" />
                  <span className="text-sm font-medium text-text">{config.name || t.mcpSettings.unnamed}</span>
                </div>
                <button
                  onClick={() => updateConfig(config.id, 'enabled', !config.enabled)}
                  className={`w-9 h-5 rounded-full transition-colors duration-150 cursor-pointer flex items-center ${
                    config.enabled ? 'bg-cta justify-end' : 'bg-border justify-start'
                  }`}
                >
                  <div className="w-4 h-4 rounded-full bg-white mx-0.5" />
                </button>
              </div>
              
              <div className="grid grid-cols-2 gap-3">
                <Input 
                  label={t.mcpSettings.name}
                  value={config.name}
                  onChange={(e) => updateConfig(config.id, 'name', e.target.value)}
                  placeholder={t.mcpSettings.namePlaceholder}
                />
                <Input 
                  label={t.mcpSettings.type}
                  value={config.type}
                  onChange={(e) => updateConfig(config.id, 'type', e.target.value)}
                  placeholder={t.mcpSettings.typePlaceholder}
                />
              </div>
              
              <Input 
                label={t.mcpSettings.endpoint}
                value={config.endpoint}
                onChange={(e) => updateConfig(config.id, 'endpoint', e.target.value)}
                placeholder={t.mcpSettings.endpointPlaceholder}
              />

              <button 
                onClick={() => removeConfig(config.id)}
                className="flex items-center gap-1.5 text-xs text-accent-rose hover:text-accent-rose/80 transition-colors cursor-pointer pt-1"
              >
                <Trash2 className="w-3 h-3" />
                {t.mcpSettings.delete}
              </button>
            </div>
          ))}
        </div>

        <Button variant="secondary" className="w-full">
          {t.mcpSettings.saveConfig}
        </Button>
      </div>
    </div>
  );
}
