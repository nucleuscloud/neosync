import Footer from '@/components/Footer';
import GithubBanner from '@/components/banner/GithubBanner';
import TopNav from '@/components/nav/TopNav';
import { Metadata } from 'next';
import Script from 'next/script';
import { Suspense } from 'react';
import '../styles/global.css';
import PHProvider, { PostHogPageview } from './providers';

export const metadata: Metadata = {
  metadataBase: new URL('https://assets.nucleuscloud.com/'),
  openGraph: {
    title: 'Neosync',
    description: 'Open Source Synthetic Data Orchestration',
    url: 'https://neosync.dev',
    siteName: 'Neosync',
    images: [
      {
        url: '/neosync/marketingsite/mainOGHero.svg',
        width: 1200,
        height: 630,
        alt: 'mainOG',
      },
    ],
    locale: 'en_US',
    type: 'website',
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <Suspense>
        <PostHogPageview />
      </Suspense>
      <PHProvider>
        <body>
          <Script
            async
            src={`https://www.googletagmanager.com/gtag/js? 
      id=${process.env.GTAG}`}
          ></Script>
          <Script
            id="google-analytics"
            dangerouslySetInnerHTML={{
              __html: `
          window.dataLayer = window.dataLayer || [];
          function gtag(){dataLayer.push(arguments);}
          gtag('js', new Date());

          gtag('config', '${process.env.GTAG}');
        `,
            }}
          ></Script>
          <div className="flex flex-col w-full relative">
            <GithubBanner />
            <TopNav />
            <div>{children}</div>
            <Footer />
          </div>
        </body>
      </PHProvider>
    </html>
  );
}
