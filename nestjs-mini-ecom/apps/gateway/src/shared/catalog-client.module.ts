import { CATALOG_PACKAGE_NAME } from '@app/proto/catalog';
import { CATALOG_CLIENT } from '@app/proto/catalog.constant';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { join } from 'path';
import type { Env } from '../config/env';
import { CatalogSvcProvider } from './clients.provider';

@Module({
  imports: [
    ClientsModule.registerAsync([
      {
        name: CATALOG_CLIENT,
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
    ]),
  ],
  providers: [CatalogSvcProvider],
  exports: [CatalogSvcProvider],
})
export class CatalogClientModule {}
