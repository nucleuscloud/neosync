import { NextResponse } from 'next/server';
import { auth, getLogoutUrl } from '../[...nextauth]/auth';

export async function GET(): Promise<NextResponse> {
  try {
    const nextauthUrl = process.env.NEXTAUTH_URL!;
    const session = await auth();
    if (!session) {
      return NextResponse.redirect(nextauthUrl);
    }

    const logoutUrl = await getLogoutUrl();
    if (!logoutUrl) {
      throw new Error('unable to locate logout url');
    }
    if (!session.idToken) {
      throw new Error('no id token present in the session');
    }

    await fetch(logoutUrl, {
      method: 'POST',
      body: new URLSearchParams({
        id_token_hint: session.idToken,
        // Needed for OAuth logout endpoint
        post_logout_redirect_uri: nextauthUrl,
      }),
    });

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error(error);
    return NextResponse.json({
      success: false,
      message: 'Could not sign out of auth provider',
    });
  }
}
