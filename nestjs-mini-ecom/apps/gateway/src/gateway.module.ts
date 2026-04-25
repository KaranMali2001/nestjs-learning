import { CATALOG_PACKAGE_NAME } from '@app/proto/catalog';
import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { join } from 'path';
import { Env, envSchema } from './config/env';
import { GatewayController } from './gateway.controller';
import { GatewayService } from './gateway.service';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: 'apps/gateway/.env',
      validate: (config) => envSchema.parse(config),
    }),
    ClientsModule.registerAsync([
      {
        name: 'CATALOG_SERVICE',
        inject: [ConfigService],
        useFactory: (config: ConfigService<Env, true>) => ({
          transport: Transport.GRPC,
          options: {
            package: CATALOG_PACKAGE_NAME,
            protoPath: join(__dirname, 'catalog.proto'),
            url: config.getOrThrow('CATALOG_GRPC_URL'),
          },
        }),
      },
      {
        name: 'MEDIA_SERVICE',
        inject: [ConfigService],
        useFactory: (config: ConfigService<Env, true>) => ({
          transport: Transport.RMQ,
          options: {
            urls: [config.getOrThrow('RABBITMQ_URL')] as [string],
            queue: config.getOrThrow('MEDIA_QUEUE_NAME'),
          },
        }),
      },
      {
        name: 'SEARCH_SERVICE',
        inject: [ConfigService],
        useFactory: (config: ConfigService<Env, true>) => ({
          transport: Transport.RMQ,
          options: {
            urls: [config.getOrThrow('RABBITMQ_URL')] as [string],
            queue: config.getOrThrow('SEARCH_QUEUE_NAME'),
          },
        }),
      },
    ]),
  ],
  controllers: [GatewayController],
  providers: [GatewayService],
})
export class GatewayModule {}
