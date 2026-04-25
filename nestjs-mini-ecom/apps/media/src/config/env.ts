import z from 'zod';

export const envSchema = z.object({
  SERVICE_NAME: z.string(),
  RABBITMQ_URL: z.string().url(),
  RABBITMQ_QUEUE: z.string(),
});
export type Env = z.infer<typeof envSchema>;
