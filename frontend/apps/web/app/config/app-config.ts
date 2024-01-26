export interface SystemAppConfig {
  isAuthEnabled: boolean;
  publicAppBaseUrl: string;
  posthog: PosthogConfig;
  isNeosyncCloud: boolean;
  isKubernetes: boolean;
}

interface PosthogConfig {
  enabled: boolean;
  key?: string;
  host: string;
}
