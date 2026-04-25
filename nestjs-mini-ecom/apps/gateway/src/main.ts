import { ConsoleLogger, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { NestFactory } from '@nestjs/core';
import { Env } from './config/env';
import { LoggingInterceptor } from './interceptors/logging.interceptor';
import { GatewayModule } from './gateway.module';

async function bootstrap() {
  const app = await NestFactory.create(GatewayModule, {
    logger: new ConsoleLogger({ colors: true }),
  });

  app.useGlobalInterceptors(new LoggingInterceptor());

  const config = app.get(ConfigService<Env>);
  const port = config.getOrThrow<number>('PORT');
  const serviceName = config.getOrThrow<string>('SERVICE_NAME');
  const logger = new Logger(serviceName);
  await app.listen(port);
  logger.log(`${serviceName} is running on port ${port}`);
}
bootstrap();
