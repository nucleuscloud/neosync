'use client';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { useNeosyncUser } from '@/libs/hooks/useNeosyncUser';
import { useSession } from 'next-auth/react';
import Script from 'next/script';
import { ReactElement, useEffect } from 'react';
import { useAccount } from './account-provider';

export function UnifyScriptProvider(): ReactElement {
  const { data: systemAppConfig, isLoading } = useGetSystemAppConfig();

  if (
    isLoading ||
    !systemAppConfig?.unify.enabled ||
    !systemAppConfig.unify.key
  ) {
    return <></>;
  }
  return (
    <Script
      id="unify"
      dangerouslySetInnerHTML={{
        __html: `!function(){window.unify||(window.unify=Object.assign([],["identify","page","startAutoPage","stopAutoPage","startAutoIdentify","stopAutoIdentify"].reduce((function(t,e){return t[e]=function(){return unify.push([e,[].slice.call(arguments)]),unify},t}),{})));var t=document.createElement("script");t.async=!0,t.setAttribute("src","https://tag.unifyintent.com/v1/3bzXn1sjuq1cb6wQF3Cp86/script.js"),t.setAttribute("data-api-key","${systemAppConfig.unify.key}"),t.setAttribute("id","unifytag"),(document.body||document.head).appendChild(t)}();`,
      }}
    />
  );
}

const isBrowser = () => typeof window !== 'undefined';

export function UnifyIdentifier(): ReactElement {
  const { data: systemAppConfig, isLoading: isSystemAppConfigLoading } =
    useGetSystemAppConfig();
  const { data: userData, isLoading: isUserDataLoading } = useNeosyncUser();
  const { data: session } = useSession();
  const { account, isLoading: isAccountLoading } = useAccount();
  const user = session?.user;

  useEffect(() => {
    if (
      !isBrowser() ||
      isUserDataLoading ||
      isAccountLoading ||
      isSystemAppConfigLoading ||
      !systemAppConfig?.unify.enabled ||
      !systemAppConfig.unify.key ||
      !systemAppConfig.isAuthEnabled ||
      !user?.email
    ) {
      return;
    }
    // we only want to set the user id if auth is enabled, otherwise it is always the same
    // so it makes it harder to identify unique posthog sessions when running in un-auth mode.
    const userId = systemAppConfig?.isAuthEnabled
      ? userData?.userId
      : undefined;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).ko?.identify(user?.email, {
      accountName: account?.name,
      accountId: account?.id,
      email: user?.email,
      name: user?.name,
      neosyncCloud: systemAppConfig?.isNeosyncCloud ?? false,
      userId,
    });
  }, [
    isUserDataLoading,
    isAccountLoading,
    isSystemAppConfigLoading,
    account?.id,
    account?.name,
    userData?.userId,
    systemAppConfig?.isAuthEnabled,
    systemAppConfig?.isNeosyncCloud,
    user?.email,
    user?.name,
    systemAppConfig?.unify.enabled,
    systemAppConfig?.unify.key,
  ]);

  return <></>;
}
