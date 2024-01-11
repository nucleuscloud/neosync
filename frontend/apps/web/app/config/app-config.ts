export interface SystemAppConfig {
  isAuthEnabled: boolean;
  publicAppBaseUrl: string;
  posthog: PosthogConfig;
}

interface PosthogConfig {
  enabled: boolean;
  key?: string;
  host: string;
}
