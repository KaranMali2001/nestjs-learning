import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import type { Env } from '../config/env';
import { SearchSvcProvider } from './clients.provider';

@Module({
  imports: [
    ClientsModule.registerAsync([
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
  providers: [SearchSvcProvider],
  exports: [SearchSvcProvider],
})
export class SearchClientModule {}
