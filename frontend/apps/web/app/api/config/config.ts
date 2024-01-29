import { SystemAppConfig } from '@/app/config/app-config';

// This will only be hydrated with env vars if invoked on the server
// Unfortunately, during a standalone build, this method is invoked and the values here are used as environment variables.
// These aren't provided at build time so will fall back to their defaults.
// This only seems to be an issue with the root layout.tsx, where as all sub pages cause a re-render of the root layout
// which causes them to be their correct values. However, if navigating to "/", the root layout isn't re-rendered and is given the defaults.
export function getSystemAppConfig(): SystemAppConfig {
  return {
    isAuthEnabled: process.env.AUTH_ENABLED === 'true',
    publicAppBaseUrl:
      process.env.NEXT_PUBLIC_APP_BASE_URL ?? 'http://localhost:3000',
    posthog: {
      enabled: process.env.NEOSYNC_ANALYTICS_ENABLED
        ? process.env.NEOSYNC_ANALYTICS_ENABLED === 'true'
        : true,
      host: process.env.POSTHOG_HOST ?? 'https://app.posthog.com',
      key: process.env.POSTHOG_KEY,
    },
    isNeosyncCloud: process.env.NEOSYNC_CLOUD === 'true',
    enableRunLogs: process.env.ENABLE_RUN_LOGS === 'true',
    signInProviderId: process.env.AUTH_PROVIDER_ID,
  };
}
