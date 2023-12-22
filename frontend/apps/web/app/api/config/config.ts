import { SystemAppConfig } from '@/app/config/app-config';

// This will only be hydrated with env vars if invoked on the server
export function getSystemAppConfig(): SystemAppConfig {
  return {
    isAuthEnabled: process.env.AUTH_ENABLED == 'true',
    publicAppBaseUrl:
      process.env.NEXT_PUBLIC_APP_BASE_URL ?? 'http://localhost:3000',
    posthog: {
      enabled: process.env.NEOSYNC_ANALYTICS_ENABLED
        ? process.env.NEOSYNC_ANALYTICS_ENABLED == 'true'
        : true,
      host: process.env.POSTHOG_HOST ?? 'https://app.posthog.com',
      key: process.env.POSTHOG_KEY,
    },
  };
}
