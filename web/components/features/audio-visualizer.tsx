'use client';

import { useRef, useEffect } from 'react';
import { cn } from '@/lib/utils';
import { useTranslation } from '@/lib/i18n';

interface AudioVisualizerProps {
  /** Getter that returns the live frequency data buffer (mutated in-place by the recorder) */
  getFrequencyData: () => Uint8Array;
  isActive: boolean;
  isPaused?: boolean;
  className?: string;
  barCount?: number;
  barColor?: string;
  barMinHeight?: number;
  barMaxHeight?: number;
  barWidth?: number;
  barGap?: number;
}

/**
 * iPhone-style audio frequency bar visualizer.
 * Uses a self-running rAF loop that calls getFrequencyData() each frame
 * to always read the latest buffer, even if the reference was swapped.
 */
export function AudioVisualizer({
  getFrequencyData,
  isActive,
  isPaused = false,
  className,
  barCount = 40,
  barColor = '#fb7185',
  barMinHeight = 2,
  barMaxHeight = 48,
  barWidth = 2.5,
  barGap = 1.5,
}: AudioVisualizerProps) {
  const t = useTranslation();
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const prevHeightsRef = useRef<number[]>([]);
  const rafRef = useRef<number>(0);

  // Keep latest props in refs so the rAF loop always reads fresh values
  const getDataRef = useRef(getFrequencyData);
  getDataRef.current = getFrequencyData;

  const activeRef = useRef(isActive);
  activeRef.current = isActive;

  const pausedRef = useRef(isPaused);
  pausedRef.current = isPaused;

  // Single rAF loop — starts once on mount, never restarts.
  useEffect(() => {
    const heights = prevHeightsRef;

    function draw() {
      const canvas = canvasRef.current;
      if (!canvas) {
        rafRef.current = requestAnimationFrame(draw);
        return;
      }

      const ctx = canvas.getContext('2d');
      if (!ctx) {
        rafRef.current = requestAnimationFrame(draw);
        return;
      }

      const dpr = window.devicePixelRatio || 1;
      const rect = canvas.getBoundingClientRect();
      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;
      ctx.scale(dpr, dpr);

      const width = rect.width;
      const height = rect.height;
      const centerY = height / 2;

      ctx.clearRect(0, 0, width, height);

      if (heights.current.length !== barCount) {
        heights.current = new Array(barCount).fill(barMinHeight);
      }

      // Always call the getter to get the latest buffer reference
      const data = getDataRef.current();
      const binCount = data.length;
      const totalBarWidth = barCount * (barWidth + barGap) - barGap;
      const startX = (width - totalBarWidth) / 2;
      const active = activeRef.current;
      const paused = pausedRef.current;

      for (let i = 0; i < barCount; i++) {
        const freqIndex = Math.floor(Math.pow(i / barCount, 1.5) * (binCount * 0.8));
        const value = data[Math.min(freqIndex, binCount - 1)] || 0;
        const normalized = value / 255;

        const centerWeight = 1 - Math.abs((i - barCount / 2) / (barCount / 2)) * 0.3;
        let targetHeight: number;

        if (!active || paused) {
          targetHeight = barMinHeight;
        } else {
          targetHeight = barMinHeight + normalized * (barMaxHeight - barMinHeight) * centerWeight;
        }

        const prev = heights.current[i];
        const springFactor = targetHeight > prev ? 0.3 : 0.15;
        const currentHeight = prev + (targetHeight - prev) * springFactor;
        heights.current[i] = currentHeight;

        const x = startX + i * (barWidth + barGap);
        const halfH = currentHeight / 2;

        const gradient = ctx.createLinearGradient(x, centerY - halfH, x, centerY + halfH);
        gradient.addColorStop(0, barColor + 'cc');
        gradient.addColorStop(0.5, barColor);
        gradient.addColorStop(1, barColor + 'cc');

        ctx.fillStyle = gradient;
        ctx.beginPath();
        ctx.roundRect(x, centerY - halfH, barWidth, currentHeight, barWidth / 2);
        ctx.fill();
      }

      if (active && !paused) {
        const avgHeight = heights.current.reduce((a, b) => a + b, 0) / barCount;
        const glowIntensity = Math.min(avgHeight / barMaxHeight, 0.4) * 0.15;
        const glow = ctx.createRadialGradient(
          width / 2, centerY, 0,
          width / 2, centerY, width / 2,
        );
        glow.addColorStop(0, barColor + Math.round(glowIntensity * 255).toString(16).padStart(2, '0'));
        glow.addColorStop(1, barColor + '00');
        ctx.fillStyle = glow;
        ctx.fillRect(0, 0, width, height);
      }

      rafRef.current = requestAnimationFrame(draw);
    }

    rafRef.current = requestAnimationFrame(draw);
    return () => cancelAnimationFrame(rafRef.current);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [barCount, barColor, barMinHeight, barMaxHeight, barWidth, barGap]);

  return (
    <canvas
      ref={canvasRef}
      className={cn('w-full', className)}
      style={{ height: barMaxHeight + 16 }}
      role="img"
      aria-label={isActive ? (isPaused ? t.audioVisualizer.paused : t.audioVisualizer.active) : t.audioVisualizer.inactive}
    />
  );
}
