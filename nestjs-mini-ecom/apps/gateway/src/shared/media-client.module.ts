import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import type { Env } from '../config/env';
import { MediaSvcProvider } from './clients.provider';

@Module({
  imports: [
    ClientsModule.registerAsync([
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
    ]),
  ],
  providers: [MediaSvcProvider],
  exports: [MediaSvcProvider],
})
export class MediaClientModule {}
