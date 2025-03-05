import { withNeosyncContext } from '@/api-only/neosync-context';
import { getSystemAppConfig } from '@/app/api/config/config';
import { ConnectError } from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

const SLACK_REDIRECT_URL = '/hooks/slack';

export async function GET(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const searchParams = req.nextUrl.searchParams;
    const code = searchParams.get('code') || '';
    const state = searchParams.get('state') || '';

    if (!code || !state) {
      const searchParams = new URLSearchParams();
      searchParams.set('error', 'missing_code_or_state');
      return NextResponse.redirect(
        new URL(`${SLACK_REDIRECT_URL}?${searchParams.toString()}`)
      );
    }

    const systemConfig = getSystemAppConfig();

    try {
      await ctx.client.accountHooks.handleSlackOAuthCallback({
        code,
        state,
      });
      return NextResponse.redirect(
        new URL(SLACK_REDIRECT_URL, systemConfig.publicAppBaseUrl)
      );
    } catch (err) {
      const searchParams = new URLSearchParams();
      searchParams.set('error', 'failed_to_handle_slack_oauth_callback');
      if (err instanceof ConnectError) {
        searchParams.set('error_message', err.message);
      } else if (err instanceof Error) {
        searchParams.set('error_message', err.message);
      } else {
        searchParams.set('error_message', `unknown error: ${err}`);
      }
      return NextResponse.redirect(
        new URL(
          `${SLACK_REDIRECT_URL}?${searchParams.toString()}`,
          systemConfig.publicAppBaseUrl
        )
      );
    }
  })(req);
}
