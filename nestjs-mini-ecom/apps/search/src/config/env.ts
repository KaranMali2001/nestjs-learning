import z from 'zod';

export const envSchema = z.object({
  SERVICE_NAME: z.string(),
  RABBITMQ_URL: z.string(),
  RABBITMQ_QUEUE: z.string(),
  ELASTICSEARCH_URL: z.string(),
});

export type Env = z.infer<typeof envSchema>;
