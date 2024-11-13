import { createHmac, timingSafeEqual } from 'crypto';
import { NextRequest, NextResponse } from 'next/server';
import * as Yup from 'yup';

const SIG_SECRET = process.env.KEYCLOAK_SLACK_WEBHOOK_HMAC_SECRET;
const KEYCLOAK_SIG_HEADER = 'X-Keycloak-Signature';
const ATTIO_API_KEY = process.env.ATTIO_API_KEY;

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
    email: Yup.string().required('The Email is required.'),
    first_name: Yup.string(),
    last_name: Yup.string(),
    identity_provider: Yup.string(),
  }).required('The Details are required.'),
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

    let peopleRecordId = '';

    const commonHeaders = {
      Authorization: `Bearer ${ATTIO_API_KEY}`,
      'Content-Type': 'application/json',
    };

    // create the person first and then use that to associate it when we create the deal otherwise it fails
    try {
      const peopleBody = {
        data: {
          values: {
            email_addresses: [registerEvent.details.email],
            name: getFullname(registerEvent),
          },
        },
      };
      const response = await fetch(
        'https://api.attio.com/v2/objects/people/records',
        {
          method: 'POST',
          headers: commonHeaders,
          body: JSON.stringify(peopleBody),
        }
      );

      if (!response.ok) {
        throw new Error('API call failed with status ' + response.statusText);
      }

      const responseData = await response.json();
      peopleRecordId = responseData.data.id.record_id;
    } catch (err) {
      return NextResponse.json(
        { message: 'unable to create person record', error: err?.toString() },
        { status: 500 }
      );
    }

    // Then try to create the deal
    try {
      const dealbody = {
        data: {
          values: {
            name: registerEvent.details.email,
            stage: 'App sign up',
            owner: 'evis@nucleuscloud.com',
            associated_people: [
              {
                target_object: 'people',
                target_record_id: peopleRecordId,
              },
            ],
          },
        },
      };
      const response = await fetch(
        'https://api.attio.com/v2/objects/deals/records/',
        {
          method: 'POST',
          headers: commonHeaders,
          body: JSON.stringify(dealbody),
        }
      );

      if (!response.ok) {
        throw new Error(
          'Deal creation failed with status ' + response.statusText
        );
      }

      return NextResponse.json({ message: 'ok', contents: JSON.parse(text) });
    } catch (err) {
      return NextResponse.json(
        { message: 'unable to create deal record', error: err?.toString() },
        { status: 500 }
      );
    }
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

interface Name {
  first_name: string;
  last_name: string;
  full_name: string;
}

function getFullname(event: RegisterEvent): Name | undefined {
  const fullname: string[] = [];
  if (event.details.first_name) {
    fullname.push(event.details.first_name);
  }
  if (event.details.last_name) {
    fullname.push(event.details.last_name);
  }
  if (fullname.length !== 2) {
    return undefined;
  }

  return {
    first_name: fullname[0],
    last_name: fullname[1],
    full_name: fullname.join(' '),
  };
}
