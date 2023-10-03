import { CubeIcon, GearIcon } from '@radix-ui/react-icons';
import { z } from 'zod';
export const tranformerSchema = z.object({
  name: z.string(),
  type: z.string(),
  createdAt: z.date(),
  updatedAt: z.date(),
});

export type Transformer = z.infer<typeof tranformerSchema>;

export const tranformerTypes = [
  {
    value: 'system',
    label: 'System',
    icon: CubeIcon,
  },
  {
    value: 'custom',
    label: 'Custom',
    icon: GearIcon,
  },
];
