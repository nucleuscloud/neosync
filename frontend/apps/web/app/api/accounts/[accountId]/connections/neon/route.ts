import { NextRequest, NextResponse } from 'next/server';
import { Issuer, generators } from 'openid-client';

export async function POST(
  req: NextRequest,
  res: NextResponse
): Promise<NextResponse> {
  const state = 'iojeowihfj289yh923h2983hf9';

  const neonIssue = await Issuer.discover(
    'https://oauth2.neon.tech/.well-known/openid-configuration'
  );

  // this needs to be stored in a session somewhere
  // const code_verifier = generators.codeVerifier();
  const code_verifier = 'thisIsTheCodeVerfiierthisIsTheCodeVerfiiers';
  const code_challenge = generators.codeChallenge(code_verifier);

  console.log('code changelle', code_challenge);

  const client = new neonIssue.Client({
    client_id: 'neosync',
    client_secret: 'xxx',
    redirect_uris: ['http://localhost:3000/api/integrations/neon/callback'],
    response_types: ['code'],
  });

  const authUrl = client.authorizationUrl({
    scope:
      'offline offline_access urn:neoncloud:projects:create urn:neoncloud:projects:read',
    code_challenge,
    code_challenge_method: 'S256',
    state: state,
  });

  return NextResponse.json({ redirect: authUrl });
}
