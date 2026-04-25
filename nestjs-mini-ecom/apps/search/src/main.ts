import { Logger } from '@nestjs/common';
import { NestFactory } from '@nestjs/core';
import { MicroserviceOptions, Transport } from '@nestjs/microservices';
import * as dotenv from 'dotenv';
import { envSchema } from './config/env';
import { SearchModule } from './search.module';

async function bootstrap() {
  dotenv.config({ path: 'apps/search/.env' });
  const transportProtocol = Transport.RMQ;
  const env = envSchema.parse(process.env);
  const app = await NestFactory.createMicroservice<MicroserviceOptions>(
    SearchModule,
    {
      transport: transportProtocol,
      options: {
        urls: [env.RABBITMQ_URL],
        queue: env.RABBITMQ_QUEUE,
        queueOptions: {
          durable: true,
        },
        noAck: false,
      },
    },
  );
  const serviceName = env.SERVICE_NAME;
  const logger = new Logger(serviceName);
  await app.listen();
  app.enableShutdownHooks();
  logger.log(`${serviceName} is listening on  ${env.RABBITMQ_URL}`);
}
bootstrap();
