import { ReactElement } from 'react';
import { CopyButton } from './CopyButton';
import { ButtonProps } from './ui/button';

interface Props {
  labelText: string;
  copyText: string;
  onHoverText: string;
  onCopiedText?: string;
  copyButtonVariant?: ButtonProps['variant'];
}
// Component that displays a resource identifier and a copy text button
export default function ResourceId(props: Props): ReactElement<any> {
  const {
    labelText,
    copyText,
    onHoverText,
    onCopiedText = 'Success!',
    copyButtonVariant = 'ghost',
  } = props;

  return (
    <div className="flex flex-row items-center gap-1">
      <h3 className="text-muted-foreground text-sm">{labelText}</h3>
      <CopyButton
        onHoverText={onHoverText}
        textToCopy={copyText}
        onCopiedText={onCopiedText}
        buttonVariant={copyButtonVariant}
      />
    </div>
  );
}
