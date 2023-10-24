import { z } from 'zod';
export const connectionSchema = z.object({
  id: z.string(),
  name: z.string(),
  createdAt: z.date(),
  category: z.string(),
  status: z.string(),
});

export type Connection = z.infer<typeof connectionSchema>;
