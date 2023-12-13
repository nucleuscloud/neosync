'use client';
import { CheckIcon, CopyIcon } from '@radix-ui/react-icons';
import { ReactElement, useEffect, useRef, useState } from 'react';
import { useCopyToClipboard } from 'usehooks-ts';
import { Button, ButtonProps } from './ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from './ui/tooltip';
import { useToast } from './ui/use-toast';

interface Props {
  onHoverText: string;
  onCopiedText: string;
  textToCopy: string;
  isDisabled?: boolean;
  buttonVariant?: ButtonProps['variant'];
}
export function CopyButton(props: Props): ReactElement {
  const { onHoverText, onCopiedText, textToCopy, isDisabled, buttonVariant } =
    props;

  const [tooltipText, setTooltipText] = useState(onHoverText);
  const [justCopied, setJustCopied] = useState(false);
  const [, copyToClipboard] = useCopyToClipboard();
  const { toast } = useToast();
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (justCopied) {
      setTooltipText(onCopiedText);
    } else {
      setTooltipText(onHoverText);
    }
  }, [justCopied, onCopiedText, onHoverText]);
  const buttonRef = useRef(null);
  const iconRef = useRef(null);

  function onClick(): void {
    copyToClipboard(textToCopy)
      .then(() => setJustCopied(true))
      .catch((err) => {
        console.error(err);
        toast({
          title: 'Unable to copy text',
          variant: 'destructive',
        });
      });
  }

  function onMouseLeave(): void {
    if (justCopied) {
      setJustCopied(false);
    }
  }

  return (
    <TooltipProvider>
      <Tooltip open={open} onOpenChange={setOpen}>
        <TooltipTrigger asChild>
          <Button
            ref={buttonRef}
            variant={buttonVariant}
            type="button"
            onClick={(e) => {
              e.preventDefault();
              onClick();
            }}
            disabled={isDisabled}
            onMouseLeave={onMouseLeave}
          >
            {justCopied ? (
              <CheckIcon className="text-green-600" ref={iconRef} />
            ) : (
              <CopyIcon ref={iconRef} />
            )}
          </Button>
        </TooltipTrigger>
        <TooltipContent
          onPointerDownOutside={(event) => {
            if (
              event.target === buttonRef.current ||
              event.target === iconRef.current
            ) {
              event.preventDefault();
            }
          }}
        >
          <p>{tooltipText}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
