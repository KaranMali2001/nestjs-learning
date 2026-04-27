import z from 'zod';

export const envSchema = z.object({
  SERVICE_NAME: z.string(),
  PORT: z.coerce.number(),
  DB_HOST: z.string(),
  DB_PORT: z.coerce.number(),
  DB_USER: z.string(),
  DB_PASSWORD: z.string(),
  DB_NAME: z.string(),
  RABBITMQ_URL: z.string(),
});
export type Env = z.infer<typeof envSchema>;
