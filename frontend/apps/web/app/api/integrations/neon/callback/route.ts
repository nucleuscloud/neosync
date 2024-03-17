import axios from 'axios';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(
  req: NextRequest,
  res: NextResponse
): Promise<NextResponse> {
  const { method } = req;

  // Check if the request method is allowed
  if (method !== 'GET') {
    return NextResponse.json(
      { message: 'Method Not Allowed' },
      { status: 405 }
    );
  }

  const url = new URL(req.url);
  const params = new URLSearchParams(url.searchParams);
  const code_verifier = 'thisIsTheCodeVerfiierthisIsTheCodeVerfiiers';
  console.log('code', params.get('code'));
  console.log('params', params);

  // // compare this with the original state valu eto ensure that the original request came from our application and not from a third party
  //TODO: figure out a way to get the original state value
  /* const state = params.get('state');
  if(state != params.get('state)){
    return NextResponse.json(
      { message: 'State verification codes do not match' },
      { status: 500 }
    );
  }
  */

  const queryParams = new URLSearchParams({
    client_id: 'neosync',
    redirect_uri: 'http://localhost:3000/api/integrations/neon/callback',
    client_secret: 'xxxx',
    grant_type: 'authorization_code',
    code_verifier,
    code: params.get('code') ?? '', // exchange for access token
  });

  try {
    const tokenResponse = await axios.post(
      `https://oauth2.neon.tech/oauth2/token`,
      queryParams.toString()
    );

    return NextResponse.json({ token: tokenResponse });
  } catch (e) {
    if (axios.isAxiosError(e)) {
      console.error('there was an error', e.response?.data);
    }
  }
  return NextResponse.json(res);
}
