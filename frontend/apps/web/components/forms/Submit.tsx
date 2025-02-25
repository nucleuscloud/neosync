import { ReactElement } from 'react';
import ButtonText from '../ButtonText';
import Spinner from '../Spinner';
import { Button } from '../ui/button';

interface Props {
  isSubmitting: boolean;
  text: string;
}

export default function Submit(props: Props): ReactElement {
  const { isSubmitting, text } = props;

  return (
    <Button type="submit" disabled={isSubmitting} className="w-full sm:w-auto">
      <ButtonText
        leftIcon={isSubmitting ? <Spinner /> : undefined}
        text={text}
      />
    </Button>
  );
}
