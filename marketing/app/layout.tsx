import Footer from '@/components/Footer';
import GithubBanner from '@/components/banner/GithubBanner';
import GithubButton from '@/components/banner/GithubButton';
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
            id="koala-snippet"
            dangerouslySetInnerHTML={{
              __html: ` 
          !function(t){if(window.ko)return;window.ko=[],["identify","track","removeListeners","open","on","off","qualify","ready"].forEach(function(t){ko[t]=function(){var n=[].slice.call(arguments);return n.unshift(t),ko.push(n),ko}});var n=document.createElement("script");n.async=!0,n.setAttribute("src","https://cdn.getkoala.com/v1/pk_4fa92236b6fe5d23fb878c88c14d209fd48e/sdk.js"),(document.body || document.head).appendChild(n)}();`,
            }}
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
            <div className="flex md:hidden lg:hidden">
              <GithubBanner />
            </div>
            <div className="hidden lg:flex">
              <GithubButton />
            </div>
            <TopNav />
            <div>{children}</div>
            <Footer />
          </div>
        </body>
      </PHProvider>
    </html>
  );
}
