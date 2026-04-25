import z from 'zod';

export const envSchema = z.object({
  SERVICE_NAME: z.string(),
  RABBITMQ_URL: z.string(),
  CATALOG_GRPC_URL: z.string(),
  MEDIA_QUEUE_NAME: z.string(),
  SEARCH_QUEUE_NAME: z.string(),
  PORT: z.coerce.number(),
});
export type Env = z.infer<typeof envSchema>;
