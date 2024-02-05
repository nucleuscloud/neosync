import { env } from '@/env';
import axios from 'axios';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest, res: NextResponse) {
  const body = await req.json();

  const email = body.email;
  const company = body.company;

  const formBody = `company=${encodeURIComponent(
    company
  )}&email=${encodeURIComponent(email)}&userGroup=${encodeURIComponent(
    'private beta'
  )}`;

  const options = {
    'Content-Type': 'application/x-www-form-urlencoded',
  };

  try {
    await axios.post(
      `https://app.loops.so/api/newsletter-form/${env.LOOPS_FORM_ID}`,
      formBody,
      { headers: options }
    );
    return NextResponse.json({ message: 'successfully create the lead' });
  } catch (e) {
    throw e;
  }
}
