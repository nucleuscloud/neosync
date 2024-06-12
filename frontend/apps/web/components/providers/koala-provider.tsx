'use client';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { useSession } from 'next-auth/react';
import Script from 'next/script';
import { ReactElement, useEffect } from 'react';

export default function KoalaIdentifier(): ReactElement {
  const { data: systemAppConfig, isLoading: isSystemAppConfigLoading } =
    useGetSystemAppConfig();
  const { data: session } = useSession();
  const user = session?.user;

  console.log('test 1');

  useEffect(() => {
    if (
      isSystemAppConfigLoading ||
      typeof window == 'undefined' ||
      !systemAppConfig?.koala.enabled
    ) {
      return;
    }

    try {
      // eslint-disable-next-line
      window.ko?.identify(user?.email ?? '', {
        name: user?.name,
        neosyncCloud: systemAppConfig?.isNeosyncCloud ?? false,
      });
      console.log('test 3');
    } catch (error) {
      console.error('Error in Koala identification:', error);
    }
  }, [user?.name, systemAppConfig?.isNeosyncCloud]);

  return <></>;
}

export function KProvider() {
  const { data: systemAppConfig } = useGetSystemAppConfig();

  if (systemAppConfig?.koala.enabled) {
    <Script
      id="koala"
      dangerouslySetInnerHTML={{
        __html: `!function(t){if(window.ko)return;window.ko=[],["identify","track","removeListeners","open","on","off","qualify","ready"].forEach(function(t){ko[t]=function(){var n=[].slice.call(arguments);return n.unshift(t),ko.push(n),ko}});var n=document.createElement("script");n.async=!0,n.setAttribute("src","https://cdn.getkoala.com/v1/pk_4fa92236b6fe5d23fb878c88c14d209fd48e/sdk.js"),(document.body || document.head).appendChild(n)}();`,
      }}
    ></Script>;
  } else {
    return <></>;
  }
}
