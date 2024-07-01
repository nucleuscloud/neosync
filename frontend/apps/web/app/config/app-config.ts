export interface SystemAppConfig {
  isAuthEnabled: boolean;
  publicAppBaseUrl: string;
  posthog: PosthogConfig;
  koala: KoalaConfig;
  isNeosyncCloud: boolean;
  enableRunLogs: boolean;
  signInProviderId?: string;
  isMetricsServiceEnabled: boolean;

  calendlyUpgradeLink: string;
  isGcpCloudStorageConnectionsEnabled: boolean;
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
