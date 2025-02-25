import NextLink from 'next/link';
import { ReactElement } from 'react';
import { GrClone } from 'react-icons/gr';
import ButtonText from './ButtonText';
import { useAccount } from './providers/account-provider';
import { Button } from './ui/button';

interface CloneConnectionProps {
  id: string;
}

export function CloneConnectionButton(
  props: CloneConnectionProps
): ReactElement<any> {
  const { id } = props;
  const { account } = useAccount();

  return (
    <NextLink href={`/${account?.name}/connections/${id}/clone`}>
      <Button>
        <ButtonText text="Clone" leftIcon={<GrClone className="mr-1" />} />
      </Button>
    </NextLink>
  );
}
