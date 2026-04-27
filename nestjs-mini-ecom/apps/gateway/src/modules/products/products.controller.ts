import { Controller, Get } from '@nestjs/common';
import { Public } from '../../decorators/public.decorator';
import { ProductService } from './products.service';

@Controller('products')
export class ProductController {
  constructor(private readonly productSvc: ProductService) {}
  @Public()
  @Get('ping')
  async pingCatalogAndMedia() {
    return await this.productSvc.pingCatalogAndMedia();
  }
}
