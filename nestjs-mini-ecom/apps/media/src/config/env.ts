import z from 'zod';

export const envSchema = z.object({
  SERVICE_NAME: z.string(),
  RABBITMQ_URL: z.string(),
  RABBITMQ_QUEUE: z.string(),
  CLOUDINARY_CLOUD_NAME: z.string(),
  CLOUDINARY_API_KEY: z.string(),
  CLOUDINARY_API_SECRET: z.string(),
});
export type Env = z.infer<typeof envSchema>;
