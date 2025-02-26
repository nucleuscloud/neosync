'use client';
import { cn } from '@/libs/utils';
import { EyeNoneIcon, EyeOpenIcon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { Button } from './ui/button';
import { Input, InputProps } from './ui/input';

export const PasswordInput = ({ className, ...props }: InputProps) => {
  const [showPassword, setShowPassword] = useState<boolean>(false);
  const disabled =
    props.value === '' || props.value === undefined || props.disabled;

  return (
    <div className="relative">
      <Input
        type={showPassword ? 'text' : 'password'}
        className={cn('hide-password-toggle pr-10', className)}
        {...props}
      />
      <Button
        type="button"
        variant="ghost"
        size="sm"
        className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
        onClick={() => setShowPassword((prev) => !prev)}
        disabled={disabled}
      >
        {showPassword && !disabled ? (
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

PasswordInput.displayName = 'PasswordInput';
