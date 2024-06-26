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
