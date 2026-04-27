import { Module } from '@nestjs/common';
import { CatalogClientModule } from '../../shared/catalog-client.module';
import { MediaClientModule } from '../../shared/media-client.module';
import { ProductController } from './products.controller';
import { ProductService } from './products.service';

@Module({
  controllers: [ProductController],
  imports: [CatalogClientModule, MediaClientModule],
  providers: [ProductService],
})
export class ProductModule {}
