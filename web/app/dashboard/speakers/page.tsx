'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { useAppStore } from '@/lib/store';
import { useTranslation } from '@/lib/i18n';
import { speakersApi } from '@/lib/api';
import { SpeakerLabelManager } from '@/components/features/speaker-label-manager';

export default function SpeakersPage() {
  const { speakers, setSpeakers } = useAppStore();
  const t = useTranslation();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    speakersApi.listSpeakers()
      .then((res) => { if (res.data) setSpeakers(res.data); })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [setSpeakers]);

  const totalProfiles = speakers.length;
  const totalSamples = speakers.reduce((acc, s) => acc + (s.sampleCount || 0), 0);
  const enrolledCount = speakers.filter(s => (s.sampleCount || 0) > 0).length;

  return (
    <div className="min-h-screen">
      <Header title={t.speakers.title} description={t.speakers.description} />

      <div className="px-8 pb-8 space-y-5">
        {/* Stats bar */}
        <div className="flex items-center gap-6 py-3 border-b border-border">
          <div className="flex items-baseline gap-1.5">
            <span className="text-lg font-semibold text-text tabular-nums">{totalProfiles}</span>
            <span className="text-xs text-text-muted">{t.speakers.statProfiles}</span>
          </div>
          <div className="w-px h-5 bg-border" />
          <div className="flex items-baseline gap-1.5">
            <span className="text-lg font-medium text-text tabular-nums">{totalSamples}</span>
            <span className="text-xs text-text-muted">{t.speakers.statSamples}</span>
          </div>
          <div className="flex items-baseline gap-1.5">
            <span className="text-lg font-medium text-cta tabular-nums">{enrolledCount}</span>
            <span className="text-xs text-text-muted">{t.speakers.statEnrolled}</span>
          </div>
        </div>

        {/* Speaker list manager */}
        <SpeakerLabelManager initialLoaded={!loading} />
      </div>
    </div>
  );
}
