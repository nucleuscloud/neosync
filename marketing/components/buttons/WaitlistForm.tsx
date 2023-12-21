'use client';
import { ReactElement, useState } from 'react';
import { FireMixpanel } from '../../lib/mixpanel';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { ArrowRightIcon, ReloadIcon } from '@radix-ui/react-icons';
import { Alert, AlertTitle } from '../ui/alert';
import { CheckCheckIcon } from 'lucide-react';

type FormStatus = 'success' | 'error' | 'invalid email' | 'null';

export default function WaitlistForm(): ReactElement {
  const [email, setEmail] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [_, setSubmittedForm] = useState<boolean>(false);
  const [formStatus, setFormStatus] = useState<FormStatus>('null');

  const handleSubmit = () => {
    FireMixpanel('submitted demo', {
      source: 'demo page',
      type: 'submitted demo',
    });

    if (!isValidEmail(email)) {
      setFormStatus('invalid email');
      return;
    }

    setIsSubmitting(true);
    const data = {
      email: email,
    };

    fetch(`/api/waitlist-signup`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    }).then(async (res) => {
      if (res.status == 200) {
        setIsSubmitting(false);
        setSubmittedForm(true);
        setFormStatus('success');
        await timeout(3000);
        setFormStatus('null');
      } else {
        setSubmittedForm(false);
        setFormStatus('error');
      }
    });
  };

  function isValidEmail(email: string): boolean {
    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    return emailRegex.test(email);
  }

  function timeout(delay: number) {
    return new Promise((res) => setTimeout(res, delay));
  }

  return (
    <div className=" flex flex-col z-40 relative">
      <div className="text-gray-400 justify-start font-satoshi">
        Sign up for updates from the Neosync community.
      </div>
      <div className="flex flex-row space-x-3 pt-8 w-full">
        <Input
          type="email"
          placeholder="Work email"
          onChange={(e) => setEmail(e.target.value)}
        />
        <Button
          type="submit"
          id="get-a-demo"
          variant="secondary"
          className="h-9"
          disabled={isSubmitting}
          onClick={handleSubmit}
        >
          {isSubmitting ? (
            <ReloadIcon className="mr-2 h-4 w-4 animate-spin justify-c" />
          ) : (
            <div className="flex flex-row items-center space-x-2 w-full justify-center">
              <div className="font-normal py-2">Join</div>
              <ArrowRightIcon />
            </div>
          )}
        </Button>
      </div>
      <div className="mt-2">
        <FormSubmissionAlert status={formStatus} />
      </div>
    </div>
  );
}

interface SubmissionProps {
  status: string;
}

function FormSubmissionAlert(props: SubmissionProps): JSX.Element {
  const { status } = props;

  if (status == 'success') {
    return (
      <div className=" w-full px-4">
        <Alert
          variant="success"
          className="flex flex-row items-center space-x-3"
        >
          <div className="text-green-800">
            <CheckCheckIcon className="h-4 w-4" />
          </div>
          <div>
            <AlertTitle>Success</AlertTitle>
          </div>
        </Alert>
      </div>
    );
  } else if (status == 'error') {
    return (
      <div className=" w-full px-4">
        <Alert
          variant="destructive"
          className="flex flex-row items-center space-x-3"
        >
          <div className="text-red-800">
            <CheckCheckIcon className="h-4 w-4" />
          </div>
          <div>
            <AlertTitle>Error: Try again</AlertTitle>
          </div>
        </Alert>
      </div>
    );
  } else if (status == 'invalid email') {
    return (
      <div className=" w-full px-4">
        <Alert
          variant="destructive"
          className="flex flex-row items-center space-x-3"
        >
          <div className="text-red-800">
            <CheckCheckIcon className="h-4 w-4" />
          </div>
          <div>
            <AlertTitle>Error: Invalid Email</AlertTitle>
          </div>
        </Alert>
      </div>
    );
  }

  return <div></div>;
}
