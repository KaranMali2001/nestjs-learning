import z from 'zod';

export const envSchema = z.object({
  SERVICE_NAME: z.string(),
  PORT: z.coerce.number(),
  RABBITMQ_URL: z.string(),
  CATALOG_GRPC_URL: z.string(),
  MEDIA_QUEUE_NAME: z.string(),
  SEARCH_QUEUE_NAME: z.string(),
  CLERK_PUBLISHABLE_KEY: z.string(),
  CLERK_SECRET_KEY: z.string(),
});
export type Env = z.infer<typeof envSchema>;
