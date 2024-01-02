import { env } from '@/env';
import axios from 'axios';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest, res: NextResponse) {
  const body = await req.json();

  const email = body.email;
  const company = body.company;

  const formData = {
    legalConsentOptions: {
      consent: {
        consentToProcess: true,
        text: 'I agree to allow Example Company to store and process my personal data.',
        communications: [
          {
            value: true,
            subscriptionTypeId: 999,
            text: 'I agree to receive marketing communications from Example Company.',
          },
        ],
      },
    },
    fields: [
      {
        objectTypeId: '0-1',
        name: 'email',
        value: email,
      },
      {
        objectTypeId: '0-2',
        name: 'company',
        value: company,
      },
    ],
  };

  console.log('the data', company);
  const options = {
    Authorization: `Bearer ${env.HUBSPOT_TOKEN}`,
  };

  console.log('calling the api');

  try {
    await axios.post(
      'https://api.hsforms.com/submissions/v3/integration/secure/submit/24034913/ebe98096-48dc-4009-b38e-8812ae9a9ea1',
      formData,
      { headers: options }
    );
    return NextResponse.json({ message: 'successfully create the lead' });
  } catch (e) {
    throw e;
  }
}
