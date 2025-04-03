'use client';
import { cn } from '@/libs/utils';
import { EyeNoneIcon, EyeOpenIcon } from '@radix-ui/react-icons';
import { useState } from 'react';
import Spinner from './Spinner';
import { Button } from './ui/button';
import { Input, InputProps } from './ui/input';

interface SecurePasswordInputProps extends Omit<InputProps, 'value'> {
  value?: string;
  maskedValue?: string;
  onRevealPassword?(): Promise<string>;
}

export const SecurePasswordInput = ({
  className,
  value,
  maskedValue = '••••••••',
  onRevealPassword,
  ...props
}: SecurePasswordInputProps) => {
  const [showPassword, setShowPassword] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [revealedValue, setRevealedValue] = useState<string>('');

  const disabled = props.disabled;

  async function handleRevealPassword(): Promise<void> {
    if (showPassword) {
      setShowPassword(false);
      return;
    }

    if (onRevealPassword) {
      setIsLoading(true);
      try {
        const password = await onRevealPassword();
        setRevealedValue(password);
        setShowPassword(true);
      } catch (error) {
        console.error('Failed to reveal password:', error);
      } finally {
        setIsLoading(false);
      }
    }
  }

  const displayValue = showPassword ? revealedValue || value : maskedValue;

  return (
    <div className="relative">
      <Input
        type="text"
        className={cn('hide-password-toggle pr-10', className)}
        value={displayValue}
        readOnly
        {...props}
      />
      <Button
        type="button"
        variant="ghost"
        size="sm"
        className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
        onClick={handleRevealPassword}
        disabled={disabled || isLoading}
      >
        {isLoading ? (
          <Spinner className="h-4 w-4" />
        ) : showPassword ? (
          <EyeOpenIcon className="h-4 w-4" aria-hidden="true" />
        ) : (
          <EyeNoneIcon className="h-4 w-4" aria-hidden="true" />
        )}
        <span className="sr-only">
          {showPassword ? 'Hide password' : 'Show password'}
        </span>
      </Button>
    </div>
  );
};

SecurePasswordInput.displayName = 'SecurePasswordInput';
