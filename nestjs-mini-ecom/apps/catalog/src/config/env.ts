import z from 'zod';

export const envSchema = z.object({
  SERVICE_NAME: z.string(),
  PORT: z.coerce.number(),
});
export type Env = z.infer<typeof envSchema>;
