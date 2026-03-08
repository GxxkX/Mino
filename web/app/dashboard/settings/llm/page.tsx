'use client';

import { useEffect, useState, useCallback } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useTranslation } from '@/lib/i18n';
import { settingsApi } from '@/lib/api';
import { Brain, Eye, EyeOff, Check, AlertCircle, Loader2, Zap, Cpu, Server } from 'lucide-react';

const PROVIDER_ICONS: Record<string, typeof Zap> = {
  openai: Zap,
  anthropic: Cpu,
  ollama: Server,
};

const PROVIDERS = [
  { id: 'openai' },
  { id: 'anthropic' },
  { id: 'ollama' },
] as const;

type ProviderID = (typeof PROVIDERS)[number]['id'];

export default function LLMSettingsPage() {
  const t = useTranslation();

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState('');
  const [showKey, setShowKey] = useState(false);

  const [provider, setProvider] = useState<ProviderID>('openai');
  const [apiKey, setApiKey] = useState('');
  const [baseUrl, setBaseUrl] = useState('');
  const [model, setModel] = useState('');
  const [embeddingModel, setEmbeddingModel] = useState('');

  // Load config for a specific provider from backend
  const loadProviderConfig = useCallback((providerId: string) => {
    settingsApi
      .getLLMConfig(providerId)
      .then((cfg) => {
        setApiKey(cfg.api_key || '');
        setBaseUrl(cfg.base_url || '');
        setModel(cfg.model || '');
        setEmbeddingModel(cfg.embedding_model || '');
      })
      .catch(() => {
        // Reset fields if provider has no saved config
        setApiKey('');
        setBaseUrl('');
        setModel('');
        setEmbeddingModel('');
      });
  }, []);

  // Load current active config on mount
  useEffect(() => {
    settingsApi
      .getLLMConfig()
      .then((cfg) => {
        setProvider((cfg.provider || 'openai') as ProviderID);
        setApiKey(cfg.api_key || '');
        setBaseUrl(cfg.base_url || '');
        setModel(cfg.model || '');
        setEmbeddingModel(cfg.embedding_model || '');
      })
      .catch(() => {
        // keep defaults
      })
      .finally(() => setLoading(false));
  }, []);

  // When user switches provider, fetch that provider's config
  function handleProviderSwitch(id: ProviderID) {
    setProvider(id);
    setShowKey(false);
    setSaved(false);
    setError('');
    loadProviderConfig(id);
  }

  async function handleSave() {
    setSaving(true);
    setError('');
    setSaved(false);
    try {
      const updated = await settingsApi.updateLLMConfig({
        provider,
        api_key: apiKey,
        base_url: baseUrl,
        model,
        embedding_model: embeddingModel,
      });
      setApiKey(updated.api_key || '');
      setBaseUrl(updated.base_url || '');
      setModel(updated.model || '');
      setEmbeddingModel(updated.embedding_model || '');
      setSaved(true);
      setTimeout(() => setSaved(false), 2500);
    } catch {
      setError(t.llmSettings.saveError);
    } finally {
      setSaving(false);
    }
  }

  function providerLabel(id: ProviderID) {
    return t.llmSettings[id] as string;
  }

  function providerDesc(id: ProviderID) {
    const map: Record<ProviderID, string> = {
      openai: t.llmSettings.openaiDesc,
      anthropic: t.llmSettings.anthropicDesc,
      ollama: t.llmSettings.ollamaDesc,
    };
    return map[id];
  }

  if (loading) {
    return (
      <div className="min-h-screen">
        <Header title={t.llmSettings.title} backHref="/dashboard/settings" />
        <div className="flex items-center justify-center py-20">
          <Loader2 className="w-5 h-5 animate-spin text-text-muted" />
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      <Header title={t.llmSettings.title} backHref="/dashboard/settings" />

      <div className="px-8 pb-8 max-w-xl space-y-8">
        {/* Provider selection */}
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Brain className="w-4 h-4 text-text-muted" />
            <p className="section-label">{t.llmSettings.provider}</p>
          </div>
          <p className="text-xs text-text-muted mb-3 pl-6">{t.llmSettings.providerDesc}</p>

          <div className="grid grid-cols-3 gap-3">
            {PROVIDERS.map((p) => {
              const Icon = PROVIDER_ICONS[p.id];
              return (
              <button
                key={p.id}
                onClick={() => handleProviderSwitch(p.id)}
                className={`relative flex flex-col items-center gap-1.5 p-4 rounded-lg border transition-all duration-150 cursor-pointer ${
                  provider === p.id
                    ? 'border-cta bg-cta/5'
                    : 'border-border hover:border-text-muted/40'
                }`}
              >
                <Icon className={`w-5 h-5 ${provider === p.id ? 'text-cta' : 'text-text-muted'}`} />
                <span className="text-sm font-medium text-text">{providerLabel(p.id)}</span>
                <span className="text-[11px] text-text-muted leading-tight text-center">
                  {providerDesc(p.id)}
                </span>
                {provider === p.id && (
                  <div className="absolute top-2 right-2 w-4 h-4 rounded-full bg-cta flex items-center justify-center">
                    <Check className="w-2.5 h-2.5 text-background" />
                  </div>
                )}
              </button>
              );
            })}
          </div>
        </div>

        {/* API Key */}
        <div>
          <p className="section-label mb-3">{t.llmSettings.currentConfig}</p>

          <div className="space-y-4">
            <div className="relative">
              <Input
                label={t.llmSettings.apiKey}
                type={showKey ? 'text' : 'password'}
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder={t.llmSettings.apiKeyPlaceholder}
              />
              <button
                type="button"
                onClick={() => setShowKey(!showKey)}
                className="absolute right-3 top-[30px] text-text-muted hover:text-text transition-colors cursor-pointer"
                aria-label={showKey ? 'Hide API key' : 'Show API key'}
              >
                {showKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>

            <Input
              label={t.llmSettings.baseUrl}
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
              placeholder={t.llmSettings.baseUrlPlaceholder}
            />
          </div>
        </div>

        {/* Model config */}
        <div>
          <p className="section-label mb-3">{t.llmSettings.modelConfig}</p>

          <div className="space-y-4">
            <Input
              label={t.llmSettings.model}
              value={model}
              onChange={(e) => setModel(e.target.value)}
              placeholder={t.llmSettings.modelPlaceholder}
            />

            <Input
              label={t.llmSettings.embeddingModel}
              value={embeddingModel}
              onChange={(e) => setEmbeddingModel(e.target.value)}
              placeholder={t.llmSettings.embeddingModelPlaceholder}
            />
          </div>
        </div>

        {/* Status messages */}
        {saved && (
          <div className="flex items-center gap-2 text-cta text-sm">
            <Check className="w-4 h-4" />
            {t.llmSettings.saved}
          </div>
        )}
        {error && (
          <div className="flex items-center gap-2 text-accent-rose text-sm">
            <AlertCircle className="w-4 h-4" />
            {error}
          </div>
        )}

        {/* Save button */}
        <Button
          className="w-full"
          onClick={handleSave}
          disabled={saving}
        >
          {saving ? (
            <>
              <Loader2 className="w-3.5 h-3.5 animate-spin" />
              {t.llmSettings.saving}
            </>
          ) : (
            t.llmSettings.save
          )}
        </Button>
      </div>
    </div>
  );
}
