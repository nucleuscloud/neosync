'use client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ArrowRightIcon, ReloadIcon } from '@radix-ui/react-icons';
import { CheckCheckIcon } from 'lucide-react';
import { ReactElement, useState } from 'react';
import { BiErrorCircle } from 'react-icons/bi';
import { FireMixpanel } from '../../lib/mixpanel';
import { Alert, AlertTitle } from '../ui/alert';

type FormStatus = 'success' | 'error' | 'invalid email' | 'null';

export default function PrivateBetaForm(): ReactElement {
  const [email, setEmail] = useState<string>('');
  const [company, setCompany] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [_, setSubmittedForm] = useState<boolean>(false);
  const [formStatus, setFormStatus] = useState<FormStatus>('null');

  const handleSubmit = () => {
    FireMixpanel('submitted private beta', {
      source: 'private beta page',
      type: 'submitted pivate beta',
    });

    if (!isValidEmail(email)) {
      setFormStatus('invalid email');
      return;
    }

    setIsSubmitting(true);
    const data = {
      email: email,
      company: company,
    };

    fetch(`/api/private-beta`, {
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
        setIsSubmitting(false);
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
    <div className=" flex flex-col z-40 w-full">
      <div className="flex flex-col gap-3 pt-8 w-full">
        <Input
          type="email"
          placeholder="Work email"
          className="bg-gray-100 text-gray-900"
          onChange={(e) => setEmail(e.target.value)}
        />
        <Input
          type="text"
          className="bg-gray-100 text-gray-900"
          placeholder="Company"
          onChange={(e) => setCompany(e.target.value)}
        />
      </div>
      <div className="pt-10">
        <Button
          type="submit"
          id="get-a-demo"
          className="h-9 w-full"
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
      <div className=" w-full">
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
      <div className=" w-full">
        <Alert
          variant="destructive"
          className="flex flex-row items-center space-x-3"
        >
          <div className="text-red-800">
            <BiErrorCircle className="h-4 w-4" />
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
