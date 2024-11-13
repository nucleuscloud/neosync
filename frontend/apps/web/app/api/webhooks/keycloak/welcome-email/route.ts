import { createHmac, timingSafeEqual } from 'crypto';
import { NextRequest, NextResponse } from 'next/server';
import * as Yup from 'yup';

const SIG_SECRET = process.env.KEYCLOAK_SLACK_WEBHOOK_HMAC_SECRET;
const KEYCLOAK_SIG_HEADER = 'X-Keycloak-Signature';
const LOOPS_API_KEY = process.env.LOOPS_API_KEY;

/*
Example register event for username/password
{
  "time": 1711651627565,
  "realmId": "7a11ef96-d5ef-4bfe-beb8-a5d81f9de464",
  "uid": "d2e0fe26-f8e6-4012-b22f-87f989c3c9e7",
  "authDetails": {
    "realmId": "neosync-stage",
    "clientId": "neosync-app",
    "userId": "88251cc6-7408-4b31-a4c7-b2840d99b916",
    "ipAddress": "111.11.111.11",
    "username": "nick@example.com",
    "sessionId": "3cd8d025-d5d4-4ff2-82d3-a9c552ba422c"
  },
  "type": "access.REGISTER",
  "details": {
    "auth_method": "openid-connect",
    "auth_type": "code",
    "register_method": "form",
    "last_name": "Zelei",
    "redirect_uri": "https://app.stage.neosync.dev/api/auth/callback/neosync",
    "first_name": "Nick",
    "code_id": "3cd8d025-d5d4-4ff2-82d3-a9c552ba422c",
    "email": "nick@example.com",
    "username": "nick@example.com"
  }
}

Example register event for google
  {
    "time": 1713472787368,
    "realmId": "7a11ef96-d5ef-4bfe-beb8-a5d81f9de464",
    "uid": "cb444835-f5c3-4666-8a6f-1adfa6c0391d",
    "authDetails": {
      "realmId": "neosync-stage",
      "clientId": "neosync-app",
      "userId": "df684d95-fb16-457d-b461-3abfed5a7780",
      "ipAddress": "111.11.111.11",
      "username": "nickzelei@example.com",
      "sessionId": "6b4ef980-feff-418d-8e24-a2d6513b3f61"
    },
    "type": "access.REGISTER",
    "details": {
      "identity_provider": "google",
      "register_method": "broker",
      "identity_provider_identity": "nickzelei@example.com",
      "code_id": "6b4ef980-feff-418d-8e24-a2d6513b3f61",
      "email": "nickzelei@example.com",
      "username": "nickzelei@example.com"
      }
  }
*/
const RegisterEvent = Yup.object({
  time: Yup.number().required('The Time is required.'),
  type: Yup.string()
    .oneOf(['access.REGISTER'])
    .required('The Type is required.'),
  authDetails: Yup.object({
    userId: Yup.string().required('The userId is required.'),
    ipAddress: Yup.string().required('The IP Address is required.'),
  }),
  details: Yup.object({
    email: Yup.string().required('The email is required.'),
    first_name: Yup.string(),
    last_name: Yup.string(),
    identity_provider: Yup.string(),
  }).required('The Details object is required.'),
});
type RegisterEvent = Yup.InferType<typeof RegisterEvent>;

// Note, when testing this method, the body must be sent in the raw, unbeautified format for the signature to work correctly
export async function POST(req: NextRequest): Promise<NextResponse> {
  if (!SIG_SECRET) {
    return NextResponse.json(
      { message: 'missing signature secret in environment' },
      { status: 500 }
    );
  }
  const incomingSignature = req.headers.get(KEYCLOAK_SIG_HEADER);
  if (!incomingSignature) {
    return NextResponse.json(
      { message: 'must provide sigure in header' },
      { status: 403 }
    );
  }

  try {
    const text = await req.text();

    const isTrusted = verifySignature(text, SIG_SECRET, incomingSignature);
    if (!isTrusted) {
      return NextResponse.json(
        {
          message:
            'the signature in the header differs from the computed request body',
        },
        { status: 403 }
      );
    }

    const registerEvent = await RegisterEvent.validate(JSON.parse(text));

    const contactBody = {
      email: registerEvent.details.email,
      firstName: registerEvent.details.first_name,
      lastName: registerEvent.details.last_name,
      userGroup: 'app-sign-ups',
    };

    const eventBody = {
      email: registerEvent.details.email,
      eventName: 'signUp',
    };

    const commonHeaders = {
      Authorization: `Bearer ${LOOPS_API_KEY}`,
      'Content-Type': 'application/json',
    };

    try {
      // create the contact
      await fetch('https://app.loops.so/api/v1/contacts/create', {
        method: 'POST',
        headers: commonHeaders,
        body: JSON.stringify(contactBody),
      });
    } catch (err) {
      return NextResponse.json(
        { message: 'unable to complete request', error: err },
        { status: 500 }
      );
    }
    // send the event to trigger the loop
    try {
      await fetch('https://app.loops.so/api/v1/events/send', {
        method: 'POST',
        headers: commonHeaders,
        body: JSON.stringify(eventBody),
      });
    } catch (err) {
      return NextResponse.json(
        { message: 'unable to complete request', error: err },
        { status: 500 }
      );
    }

    return NextResponse.json({ message: 'ok', contents: JSON.parse(text) });
  } catch (err) {
    return NextResponse.json(
      { message: 'unable to complete request', error: err },
      { status: 500 }
    );
  }
}

function verifySignature(
  body: string,
  secret: string,
  untrustedSignature: string
): boolean {
  const signature = createHmac('sha256', secret).update(body).digest('hex');
  const trusted = new Uint8Array(Buffer.from(signature, 'ascii'));
  const untrusted = new Uint8Array(Buffer.from(untrustedSignature, 'ascii'));
  return timingSafeEqual(trusted, untrusted);
}
