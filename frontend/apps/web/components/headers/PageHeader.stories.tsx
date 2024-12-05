import { LightningBoltIcon } from '@radix-ui/react-icons';
import { action } from '@storybook/addon-actions';
import { Meta, StoryObj } from '@storybook/react';
import { expect, userEvent, within } from '@storybook/test';
import ButtonText from '../ButtonText';
import { Badge } from '../ui/badge';
import { Button } from '../ui/button';
import PageHeader from './PageHeader';

const meta: Meta<typeof PageHeader> = {
  title: 'Components/PageHeader',
  component: PageHeader,
};

export default meta;

type Story = StoryObj<typeof PageHeader>;

export const Default: Story = {
  args: {
    header: 'Default Header',
  },
};

export const FullExample: Story = {
  args: {
    header: 'PageHeader',
    leftBadgeValue: 'Important',
    extraHeading: (
      <div className="md:flex grid grid-cols-2 md:flex-row gap-4">
        <Button onClick={action('button-click')}>
          <ButtonText leftIcon={<LightningBoltIcon />} text="Button" />
        </Button>
      </div>
    ),
    leftIcon: <span>📘</span>,
    progressSteps: (
      <div className="flex gap-2">
        <Badge variant="secondary">Step 1</Badge>
        <Badge variant="secondary">Step 2</Badge>
        <Badge variant="default">Step 3</Badge>
      </div>
    ),
    pageHeaderContainerClassName: 'bg-gray-100 p-4',
    subHeadings: [
      <h3 key="1" className="text-muted-foreground text-sm">
        This is a subheading with more details
      </h3>,
    ],
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(
      canvas.getByRole('heading', { name: 'PageHeader', level: 1 })
    ).toBeInTheDocument();
    await userEvent.click(canvas.getByRole('button', { name: 'Button' }));
    await expect(canvas.getByRole('heading', { level: 3 })).toBeInTheDocument();
    await expect(canvas.getByText('Important')).toBeInTheDocument();
  },
};
