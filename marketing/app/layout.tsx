import Footer from '@/components/Footer';
import GithubBanner from '@/components/banner/GithubBanner';
import TopNav from '@/components/nav/TopNav';
import { Metadata } from 'next';
import Script from 'next/script';
import '../styles/global.css';

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

        <Script type="text/javascript" id="posthog">
          {`
!function(t,e){var o,n,p,r;e.__SV||(window.posthog=e,e._i=[],e.init=function(i,s,a){function g(t,e){var o=e.split(".");2==o.length&&(t=t[o[0]],e=o[1]),t[e]=function(){t.push([e].concat(Array.prototype.slice.call(arguments,0)))}}(p=t.createElement("script")).type="text/javascript",p.async=!0,p.src=s.api_host+"/static/array.js",(r=t.getElementsByTagName("script")[0]).parentNode.insertBefore(p,r);var u=e;for(void 0!==a?u=e[a]=[]:a="posthog",u.people=u.people||[],u.toString=function(t){var e="posthog";return"posthog"!==a&&(e+="."+a),t||(e+=" (stub)"),e},u.people.toString=function(){return u.toString(1)+".people (stub)"},o="capture identify alias people.set people.set_once set_config register register_once unregister opt_out_capturing has_opted_out_capturing opt_in_capturing reset isFeatureEnabled onFeatureFlags getFeatureFlag getFeatureFlagPayload reloadFeatureFlags group updateEarlyAccessFeatureEnrollment getEarlyAccessFeatures getActiveMatchingSurveys getSurveys onSessionId".split(" "),n=0;n<o.length;n++)g(u,o[n]);e._i.push([i,s,a])},e.__SV=1)}(document,window.posthog||[]);
posthog.init('${process.env.NEXT_PUBLIC_POSTHOG_KEY}',{api_host:'https://app.posthog.com'})

`}
        </Script>
        <div className="flex flex-col w-full relative">
          <GithubBanner />
          <TopNav />
          <div>{children}</div>
          <Footer />
        </div>
      </body>
    </html>
  );
}
