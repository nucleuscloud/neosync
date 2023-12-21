export interface SystemAppConfig {
  isAuthEnabled: boolean;
  publicAppBaseUrl: string;
  posthog: PosthogConfig;
}

export interface PosthogConfig {
  enabled: boolean;
  key?: string;
  host: string;
}
