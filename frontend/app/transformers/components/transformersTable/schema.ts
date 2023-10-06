import { CubeIcon, GearIcon } from '@radix-ui/react-icons';
import { z } from 'zod';
export const transformerSchema = z.object({
  name: z.string(),
  type: z.string(),
  description: z.string(),
});

export type Transformer = z.infer<typeof transformerSchema>;

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
