import type { Meta, StoryObj } from '@storybook/react';
import { expect, userEvent, within } from '@storybook/test';

import { PasswordInput } from './PasswordComponent';

const meta: Meta<typeof PasswordInput> = {
  title: 'Components/PasswordInput',
  component: PasswordInput,
};

export default meta;

type Story = StoryObj<typeof PasswordInput>;

export const Default: Story = {
  args: {},
};

export const HasPassword: Story = {
  args: {
    value: 'password',
  },
};

export const Disabled: Story = {
  args: {
    disabled: true,
    value: 'password',
  },
};

export const FullExample: Story = {
  args: {},
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.type(
      canvas.getByLabelText('password', { selector: 'input' }),
      'topsecret'
    );
    await expect(canvas.getByDisplayValue('topsecret')).toBeInTheDocument();

    await userEvent.click(canvas.getByRole('button'));
    await expect(canvas.getByRole('textbox')).toBeInTheDocument();
    await userEvent.click(canvas.getByRole('button'));
    await expect(canvas.queryByRole('textbox')).not.toBeInTheDocument();
  },
};
