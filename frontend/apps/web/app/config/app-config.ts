export interface SystemAppConfig {
  isAuthEnabled: boolean;
  publicAppBaseUrl: string;
  posthog: PosthogConfig;
  isNeosyncCloud: boolean;
  enableRunLogs: boolean;
  signInProviderId?: string;
}

interface PosthogConfig {
  enabled: boolean;
  key?: string;
  host: string;
}
