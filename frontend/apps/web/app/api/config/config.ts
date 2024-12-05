import { SystemAppConfig } from '@/app/config/app-config';

export const PUBLIC_PATHNAME = '/api/neosync';

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
      enabled: isAnalyticsEnabled(),
      host: process.env.POSTHOG_HOST ?? 'https://app.posthog.com',
      key: process.env.POSTHOG_KEY,
    },
    koala: {
      enabled: isAnalyticsEnabled() && !!process.env.KOALA_KEY,
      key: process.env.KOALA_KEY,
    },
    isNeosyncCloud: process.env.NEOSYNC_CLOUD === 'true',
    isStripeEnabled: process.env.STRIPE_ENABLED === 'true',
    enableRunLogs: process.env.ENABLE_RUN_LOGS === 'true',
    signInProviderId: process.env.AUTH_PROVIDER_ID,
    isMetricsServiceEnabled: process.env.METRICS_SERVICE_ENABLED === 'true',
    calendlyUpgradeLink:
      process.env.CALENDLY_UPGRADE_LINK ?? 'https://calendly.com/evis1/30min',
    isGcpCloudStorageConnectionsEnabled: isGcpConnectionsEnabled(),
    isDynamoDbConnectionsEnabled: isDynamoConnectionsEnabled(),
    isMsSqlServerEnabled: isMssqlConnectionsEnabled(),
    neosyncApiBaseUrl:
      process.env.NEOSYNC_API_BASE_URL ?? 'http://localhost:8080',
    publicNeosyncApiBaseUrl: PUBLIC_PATHNAME, // ensures that this always points to the same domain
    isJobHooksEnabled: process.env.JOBHOOKS_ENABLED === 'true',
  };
}

function isGcpConnectionsEnabled(): boolean {
  const val = process.env.GCP_CS_CONNECTIONS_DISABLED;
  return val ? val === 'false' : true;
}

function isDynamoConnectionsEnabled(): boolean {
  const val = process.env.DYNAMODB_CONNECTIONS_DISABLED;
  return val ? val === 'false' : true;
}

function isMssqlConnectionsEnabled(): boolean {
  const val = process.env.MSSQL_CONNECTIONS_DISABLED;
  return val ? val === 'false' : true;
}

function isAnalyticsEnabled(): boolean {
  const val = process.env.NEOSYNC_ANALYTICS_ENABLED;
  return val ? val === 'true' : true;
}
