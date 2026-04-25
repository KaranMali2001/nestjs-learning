import { Logger } from '@nestjs/common';
import { NestFactory } from '@nestjs/core';
import { MicroserviceOptions, Transport } from '@nestjs/microservices';
import * as dotenv from 'dotenv';
import { join } from 'path';
import { CatalogModule } from './catalog.module';
import { envSchema } from './config/env';
async function bootstrap() {
  dotenv.config({ path: 'apps/catalog/.env' });
  const transportProtocol = Transport.GRPC;
  const env = envSchema.parse(process.env);
  const packagename = env.SERVICE_NAME;
  const app = await NestFactory.createMicroservice<MicroserviceOptions>(
    CatalogModule,
    {
      transport: transportProtocol,
      options: {
        package: packagename,
        protoPath: join(__dirname, `${env.SERVICE_NAME}.proto`),
        url: `0.0.0.0:${env.PORT}`,
      },
    },
  );
  const serviceName = env.SERVICE_NAME;
  const logger = new Logger(serviceName);
  await app.listen();
  app.enableShutdownHooks();
  logger.log(`${serviceName} is using GRPC using package  ${env.SERVICE_NAME}`);
}
bootstrap();
