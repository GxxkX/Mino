import './globals.css';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Mino - AI Personal Assistant',
  description: 'Privacy-first AI personal assistant that transforms voice into structured digital assets',
  icons: {
    icon: '/logo.png',
    apple: '/logo.png',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN">
      <body className="min-h-screen bg-background">
        {children}
      </body>
    </html>
  );
}
