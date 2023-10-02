import { CheckCircledIcon, CrossCircledIcon } from '@radix-ui/react-icons';
import { z } from 'zod';
export const connectionSchema = z.object({
  id: z.string(),
  name: z.string(),
  createdAt: z.date(),
  category: z.string(),
  status: z.string(),
});

export type Connection = z.infer<typeof connectionSchema>;

export const statuses = [
  {
    value: 'disconnected',
    label: 'D/C',
    icon: CrossCircledIcon,
  },
  {
    value: 'connected',
    label: 'Connected',
    icon: CheckCircledIcon,
  },
];
