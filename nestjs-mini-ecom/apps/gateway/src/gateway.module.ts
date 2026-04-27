import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { envSchema } from './config/env';
import { ClerkAuthGuard } from './guards/clerk-auth.guard';
import { RoleGuard } from './guards/roles.guard';

import { ProductModule } from './modules/products/products.module';
import { SearchModule } from './modules/search/search.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: 'apps/gateway/.env',
      validate: (config) => envSchema.parse(config),
    }),

    ProductModule,
    SearchModule,
  ],
  providers: [RoleGuard, ClerkAuthGuard],
})
export class GatewayModule {}
