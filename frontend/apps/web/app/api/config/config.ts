import { SystemAppConfig } from '@/app/config/app-config';

export function getSystemAppConfig(): SystemAppConfig {
  return {
    isAuthEnabled: process.env.AUTH_ENABLED == 'true',
    publicAppBaseUrl:
      process.env.NEXT_PUBLIC_APP_BASE_URL ?? 'http://localhost:3000',
  };
}
