import type { Meta, StoryObj } from '@storybook/react';
import ConfirmationDialog from './DeleteConfirmationDialog';
// import { Button } from './ui/button';

const meta: Meta<typeof ConfirmationDialog> = {
  title: 'Components/ConfirmationDialog',
  component: ConfirmationDialog,
  // argTypes: {
  //   onConfirm: { action: 'confirmed' },
  //   trigger: { Control: 'none' },
  //   buttonIcon: { Control: 'none' },
  // },
};

export default meta;

type Story = StoryObj<typeof ConfirmationDialog>;

export const Default: Story = {
  args: {
    headerText: 'Are you sure?',
    description: 'This will confirm the action that you selected.',
    buttonText: 'Confirm',
    onConfirm: () => {},
  },
};

export const WithCustomTrigger: Story = {
  args: {
    // trigger: <Button variant="outline">Open Confirmation Dialog</Button>,
    headerText: 'Custom Trigger',
    description: 'This dialog was opened using a custom trigger button.',
    buttonText: 'Proceed',
    onConfirm: () => {},
  },
};

// export const DestructiveAction: Story = {
//   args: {
//     trigger: <Button variant="destructive">Delete</Button>,
//     headerText: 'Delete Item',
//     description:
//       'Are you sure you want to delete this item? This action cannot be undone.',
//     buttonText: 'Delete',
//     buttonVariant: 'destructive',
//     buttonIcon: <FiTrash2 />,
//     onConfirm: () => {},
//   },
// };
