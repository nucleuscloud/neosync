export interface SystemAppConfig {
  isAuthEnabled: boolean;
  publicAppBaseUrl: string;
  posthog: PosthogConfig;
  koala: KoalaConfig;
  isNeosyncCloud: boolean;
  isStripeEnabled: boolean;
  enableRunLogs: boolean;
  signInProviderId?: string;
  isMetricsServiceEnabled: boolean;
  isJobHooksEnabled: boolean;

  calendlyUpgradeLink: string;
  isGcpCloudStorageConnectionsEnabled: boolean;
  isDynamoDbConnectionsEnabled: boolean;
  isMsSqlServerEnabled: boolean;
  // server-side base url
  neosyncApiBaseUrl: string;
  // public (client-side) base url;
  publicNeosyncApiBaseUrl: string;
}

interface PosthogConfig {
  enabled: boolean;
  key?: string;
  host: string;
}

interface KoalaConfig {
  enabled: boolean;
  key?: string;
}
