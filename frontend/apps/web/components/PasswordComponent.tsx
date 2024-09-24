'use client';
import { cn } from '@/libs/utils';
import { EyeNoneIcon, EyeOpenIcon } from '@radix-ui/react-icons';
import * as React from 'react';
import { useState } from 'react';
import { Button } from './ui/button';
import { Input, InputProps } from './ui/input';

export const PasswordInput = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, ...props }, ref) => {
    const [showPassword, setShowPassword] = useState<boolean>(false);
    return (
      <div className="relative">
        <Input
          type={showPassword ? 'text' : 'password'}
          aria-label="password"
          aria-labelledby="password"
          className={cn('hide-password-toggle pr-10', className)}
          ref={ref}
          {...props}
        />
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
          onClick={() => setShowPassword((prev) => !prev)}
          disabled={props.disabled}
        >
          {showPassword && !props.disabled ? (
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
  }
);

PasswordInput.displayName = 'PasswordInput';
