'use client';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { useSession } from 'next-auth/react';
import { useRouter } from 'next/navigation';

export default function NeonForm() {
  const { account } = useAccount();
  const router = useRouter();
  const session = useSession();

  async function onSubmit() {
    if (!account) {
      return;
    }
    try {
      const res = await CreateNeonIntegration(account.id);
      if (res.redirect) {
        // open up in a  new window that is smaller than the current winow and ideally close when complete
        router.replace(res.redirect);
        return;
      }

      console.log('done');
    } catch (err) {
      console.error('Error in form submission:', err);
    }
  }
  // the UI
  // projects -> branches -> databases -> role -> host
  /*
  1. get all projects and list all projects
  2. user selects a project 
  3. get all branches for that project, list all branches 
  4. user selects a branch
  5. get all databases, user selects a database
  6. get all roles, list all roles 
  7. user selects a role 
  8. get role password from endpoint, get host from endpoint, connect to database, run basic query to validate connections 
  9. user clicks submit

  */

  return (
    <div>
      <Button onClick={onSubmit}>Connect your Neon account</Button>
    </div>
  );
}

async function CreateNeonIntegration(accountId: string) {
  const res = await fetch(`/api/accounts/${accountId}/connections/neon`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }

  const data = await res.json();

  return data;
}
