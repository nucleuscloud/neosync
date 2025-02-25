'use client';
import { getErrorMessage } from '@/util/util';
import { CheckIcon, CopyIcon } from '@radix-ui/react-icons';
import { ReactElement, useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';
import { useCopyToClipboard } from 'usehooks-ts';
import ButtonText from './ButtonText';
import { Button, ButtonProps } from './ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from './ui/tooltip';

interface Props {
  onHoverText: string;
  onCopiedText: string;
  textToCopy: string;
  isDisabled?: boolean;
  buttonVariant?: ButtonProps['variant'];
  buttonText?: string;
}
export function CopyButton(props: Props): ReactElement {
  const {
    onHoverText,
    onCopiedText,
    textToCopy,
    isDisabled,
    buttonVariant,
    buttonText,
  } = props;

  const [tooltipText, setTooltipText] = useState(onHoverText);
  const [justCopied, setJustCopied] = useState(false);
  const [, copyToClipboard] = useCopyToClipboard();
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
        toast.error('Unable to copy text', {
          description: getErrorMessage(err),
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
            {!!buttonText && (
              <ButtonText
                leftIcon={
                  justCopied ? (
                    <CheckIcon className="text-green-600" ref={iconRef} />
                  ) : (
                    <CopyIcon ref={iconRef} />
                  )
                }
                text={buttonText}
              />
            )}
            {!buttonText &&
              (justCopied ? (
                <CheckIcon className="text-green-600" ref={iconRef} />
              ) : (
                <CopyIcon ref={iconRef} />
              ))}
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
