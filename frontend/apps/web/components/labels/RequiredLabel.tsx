import { ReactElement } from 'react';
import { Label } from '../ui/label';

interface Props {}

export default function RequiredLabel(props: Props): ReactElement {
  const {} = props;
  return <Label className="text-red-400">* </Label>;
}
