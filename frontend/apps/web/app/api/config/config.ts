import { SystemAppConfig } from '@/app/config/app-config';

export const PUBLIC_PATHNAME = '/api/neosync';

// This will only be hydrated with env vars if invoked on the server
// Unfortunately, during a standalone build, this method is invoked and the values here are used as environment variables.
// These aren't provided at build time so will fall back to their defaults.
// This only seems to be an issue with the root layout.tsx, where as all sub pages cause a re-render of the root layout
// which causes them to be their correct values. However, if navigating to "/", the root layout isn't re-rendered and is given the defaults.
export function getSystemAppConfig(): SystemAppConfig {
  const isNeosyncCloud = process.env.NEOSYNC_CLOUD === 'true';
  return {
    isAuthEnabled: process.env.AUTH_ENABLED === 'true',
    publicAppBaseUrl:
      process.env.NEXT_PUBLIC_APP_BASE_URL ?? 'http://localhost:3000',
    posthog: {
      enabled: isAnalyticsEnabled(),
      host: process.env.POSTHOG_HOST ?? 'https://app.posthog.com',
      key: process.env.POSTHOG_KEY,
    },
    unify: {
      enabled: isAnalyticsEnabled() && !!process.env.UNIFY_KEY,
      key: process.env.UNIFY_KEY,
    },
    isNeosyncCloud,
    isStripeEnabled: process.env.STRIPE_ENABLED === 'true',
    enableRunLogs: process.env.ENABLE_RUN_LOGS === 'true',
    signInProviderId: process.env.AUTH_PROVIDER_ID,
    isMetricsServiceEnabled: process.env.METRICS_SERVICE_ENABLED === 'true',
    calendlyUpgradeLink:
      process.env.CALENDLY_UPGRADE_LINK ?? 'https://calendly.com/evis1/30min',
    isGcpCloudStorageConnectionsEnabled: isGcpConnectionsEnabled(),
    neosyncApiBaseUrl:
      process.env.NEOSYNC_API_BASE_URL ?? 'http://localhost:8080',
    publicNeosyncApiBaseUrl: PUBLIC_PATHNAME, // ensures that this always points to the same domain
    isJobHooksEnabled: process.env.JOBHOOKS_ENABLED === 'true',
    isAccountHooksEnabled:
      isNeosyncCloud || process.env.ACCOUNT_HOOKS_ENABLED === 'true',
    isSlackAccountHookEnabled:
      process.env.SLACK_ACCOUNT_HOOKS_ENABLED === 'true',
    isRbacEnabled: isNeosyncCloud || process.env.RBAC_ENABLED === 'true',
    gtag: {
      enabled: isAnalyticsEnabled() && !!process.env.GTAG,
      key: process.env.GTAG,
      conversion: process.env.GTAG_CONVERSION,
    },
  };
}

function isGcpConnectionsEnabled(): boolean {
  const val = process.env.GCP_CS_CONNECTIONS_DISABLED;
  return val ? val === 'false' : true;
}

function isAnalyticsEnabled(): boolean {
  const val = process.env.NEOSYNC_ANALYTICS_ENABLED;
  return val ? val === 'true' : true;
}
