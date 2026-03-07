'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useAppStore } from '@/lib/store';
import { useTranslation } from '@/lib/i18n';
import { settingsApi } from '@/lib/api';
import type { CloudConfig } from '@/lib/api/settings';
import {
  Cloud, Check, AlertCircle, Loader2,
  HardDrive, Database, Server, Search,
} from 'lucide-react';

export default function CloudSettingsPage() {
  const { settings, setSettings } = useAppStore();
  const t = useTranslation();

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState('');

  // Form state
  const [cfg, setCfg] = useState<CloudConfig>({
    minio_endpoint: '', minio_access_key: '', minio_secret_key: '',
    minio_secure: false, minio_region: '', minio_public_url: '',
    db_host: '', db_port: '', db_name: '', db_user: '', db_password: '', db_ssl_mode: '',
    redis_host: '', redis_port: '', redis_password: '', redis_db: 0,
    milvus_host: '', milvus_port: '', milvus_user: '', milvus_password: '', milvus_db_name: '',
    typesense_host: '', typesense_port: '', typesense_api_key: '',
  });

  useEffect(() => {
    settingsApi
      .getCloudConfig()
      .then((data) => setCfg(data))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  function update<K extends keyof CloudConfig>(key: K, value: CloudConfig[K]) {
    setCfg((prev) => ({ ...prev, [key]: value }));
  }

  async function handleSave() {
    setSaving(true);
    setError('');
    setSaved(false);
    try {
      const updated = await settingsApi.updateCloudConfig(cfg);
      setCfg(updated);
      setSaved(true);
      setTimeout(() => setSaved(false), 2500);
    } catch {
      setError(t.cloudSettings.saveError);
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen">
        <Header title={t.cloudSettings.title} backHref="/dashboard/settings" />
        <div className="flex items-center justify-center py-20">
          <Loader2 className="w-5 h-5 animate-spin text-text-muted" />
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      <Header title={t.cloudSettings.title} backHref="/dashboard/settings" />

      <div className="px-8 pb-8 max-w-xl space-y-8">
        {/* Cloud sync toggle */}
        <div className="flex items-center justify-between py-4 border-b border-border">
          <div className="flex items-center gap-3">
            <Cloud className="w-4 h-4 text-text-muted" />
            <div>
              <h3 className="text-sm font-medium text-text">{t.cloudSettings.syncTitle}</h3>
              <p className="text-xs text-text-muted mt-0.5">{t.cloudSettings.syncDesc}</p>
            </div>
          </div>
          <button
            onClick={() => setSettings({ cloudSync: !settings.cloudSync })}
            className={`w-9 h-5 rounded-full transition-colors duration-150 cursor-pointer flex items-center ${
              settings.cloudSync ? 'bg-cta justify-end' : 'bg-border justify-start'
            }`}
          >
            <div className="w-4 h-4 rounded-full bg-white mx-0.5" />
          </button>
        </div>

        {/* MinIO */}
        <div>
          <div className="flex items-center gap-2 mb-3">
            <HardDrive className="w-4 h-4 text-text-muted" />
            <p className="section-label">{t.cloudSettings.storageConfig}</p>
          </div>
          <div className="space-y-3">
            <Input
              label={t.cloudSettings.minioEndpoint}
              value={cfg.minio_endpoint}
              onChange={(e) => update('minio_endpoint', e.target.value)}
              placeholder={t.cloudSettings.minioEndpointPlaceholder}
            />
            <div className="grid grid-cols-2 gap-3">
              <Input
                label={t.cloudSettings.minioAccessKey}
                type="password"
                value={cfg.minio_access_key}
                onChange={(e) => update('minio_access_key', e.target.value)}
              />
              <Input
                label={t.cloudSettings.minioSecretKey}
                type="password"
                value={cfg.minio_secret_key}
                onChange={(e) => update('minio_secret_key', e.target.value)}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <Input
                label={t.cloudSettings.minioRegion}
                value={cfg.minio_region}
                onChange={(e) => update('minio_region', e.target.value)}
                placeholder={t.cloudSettings.minioRegionPlaceholder}
              />
              <div className="flex items-end gap-2 pb-0.5">
                <label className="flex items-center gap-2 cursor-pointer text-sm text-text">
                  <input
                    type="checkbox"
                    checked={cfg.minio_secure}
                    onChange={(e) => update('minio_secure', e.target.checked)}
                    className="accent-cta cursor-pointer"
                  />
                  {t.cloudSettings.minioSecure}
                </label>
              </div>
            </div>
            <Input
              label={t.cloudSettings.minioPublicUrl}
              value={cfg.minio_public_url}
              onChange={(e) => update('minio_public_url', e.target.value)}
              placeholder={t.cloudSettings.minioPublicUrlPlaceholder}
            />
          </div>
        </div>

        {/* PostgreSQL */}
        <div>
          <div className="flex items-center gap-2 mb-3">
            <Database className="w-4 h-4 text-text-muted" />
            <p className="section-label">{t.cloudSettings.postgresql}</p>
          </div>
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <Input
                label={t.cloudSettings.dbHost}
                value={cfg.db_host}
                onChange={(e) => update('db_host', e.target.value)}
              />
              <Input
                label={t.cloudSettings.dbPort}
                value={cfg.db_port}
                onChange={(e) => update('db_port', e.target.value)}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <Input
                label={t.cloudSettings.dbName}
                value={cfg.db_name}
                onChange={(e) => update('db_name', e.target.value)}
              />
              <Input
                label={t.cloudSettings.dbSslMode}
                value={cfg.db_ssl_mode}
                onChange={(e) => update('db_ssl_mode', e.target.value)}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <Input
                label={t.cloudSettings.dbUser}
                value={cfg.db_user}
                onChange={(e) => update('db_user', e.target.value)}
              />
              <Input
                label={t.cloudSettings.dbPassword}
                type="password"
                value={cfg.db_password}
                onChange={(e) => update('db_password', e.target.value)}
              />
            </div>
          </div>
        </div>

        {/* Redis */}
        <div>
          <div className="flex items-center gap-2 mb-3">
            <Server className="w-4 h-4 text-text-muted" />
            <p className="section-label">{t.cloudSettings.redis}</p>
          </div>
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <Input
                label={t.cloudSettings.redisHost}
                value={cfg.redis_host}
                onChange={(e) => update('redis_host', e.target.value)}
              />
              <Input
                label={t.cloudSettings.redisPort}
                value={cfg.redis_port}
                onChange={(e) => update('redis_port', e.target.value)}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <Input
                label={t.cloudSettings.redisPassword}
                type="password"
                value={cfg.redis_password}
                onChange={(e) => update('redis_password', e.target.value)}
              />
              <Input
                label={t.cloudSettings.redisDb}
                type="number"
                value={String(cfg.redis_db)}
                onChange={(e) => update('redis_db', parseInt(e.target.value) || 0)}
              />
            </div>
          </div>
        </div>

        {/* Milvus */}
        <div>
          <div className="flex items-center gap-2 mb-3">
            <Database className="w-4 h-4 text-text-muted" />
            <p className="section-label">{t.cloudSettings.milvus}</p>
          </div>
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <Input
                label={t.cloudSettings.milvusHost}
                value={cfg.milvus_host}
                onChange={(e) => update('milvus_host', e.target.value)}
              />
              <Input
                label={t.cloudSettings.milvusPort}
                value={cfg.milvus_port}
                onChange={(e) => update('milvus_port', e.target.value)}
              />
            </div>
            <Input
              label={t.cloudSettings.milvusDbName}
              value={cfg.milvus_db_name}
              onChange={(e) => update('milvus_db_name', e.target.value)}
            />
            <div className="grid grid-cols-2 gap-3">
              <Input
                label={t.cloudSettings.milvusUser}
                value={cfg.milvus_user}
                onChange={(e) => update('milvus_user', e.target.value)}
              />
              <Input
                label={t.cloudSettings.milvusPassword}
                type="password"
                value={cfg.milvus_password}
                onChange={(e) => update('milvus_password', e.target.value)}
              />
            </div>
          </div>
        </div>

        {/* Typesense */}
        <div>
          <div className="flex items-center gap-2 mb-3">
            <Search className="w-4 h-4 text-text-muted" />
            <p className="section-label">{t.cloudSettings.typesense}</p>
          </div>
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <Input
                label={t.cloudSettings.typesenseHost}
                value={cfg.typesense_host}
                onChange={(e) => update('typesense_host', e.target.value)}
              />
              <Input
                label={t.cloudSettings.typesensePort}
                value={cfg.typesense_port}
                onChange={(e) => update('typesense_port', e.target.value)}
              />
            </div>
            <Input
              label={t.cloudSettings.typesenseApiKey}
              type="password"
              value={cfg.typesense_api_key}
              onChange={(e) => update('typesense_api_key', e.target.value)}
            />
          </div>
        </div>

        {/* Status messages */}
        {saved && (
          <div className="flex items-center gap-2 text-cta text-sm">
            <Check className="w-4 h-4" />
            {t.cloudSettings.saved}
          </div>
        )}
        {error && (
          <div className="flex items-center gap-2 text-accent-rose text-sm">
            <AlertCircle className="w-4 h-4" />
            {error}
          </div>
        )}

        {/* Action buttons */}
        <div className="flex gap-3 pt-2">
          <Button
            className="flex-1"
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? (
              <>
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                {t.cloudSettings.saving}
              </>
            ) : (
              t.cloudSettings.saveConfig
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}
