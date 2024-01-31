import { NextRequest, NextResponse } from 'next/server';
import { auth, getLogoutUrl } from '../auth/[...nextauth]/auth';

export async function GET(req: NextRequest): Promise<NextResponse> {
  const nextauthUrl = process.env.NEXTAUTH_URL!;
  try {
    const { searchParams } = new URL(req.url);
    const session = await auth();
    const idToken = session?.idToken ?? searchParams.get('idToken');
    if (!idToken) {
      console.error('there was no auth session');
      return NextResponse.redirect(nextauthUrl);
    }

    const logoutUrl = await getLogoutUrl();
    if (!logoutUrl) {
      throw new Error('unable to locate logout url');
    }

    const qp = new URLSearchParams({
      id_token_hint: idToken,
      // Needed for OAuth logout endpoint
      post_logout_redirect_uri: nextauthUrl,
    });
    return NextResponse.redirect(`${logoutUrl}?${qp.toString()}`);
  } catch (error) {
    console.error('unable to redirect to provider logout', 'error: ', error);
    return NextResponse.redirect(nextauthUrl);
  }
}
