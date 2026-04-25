import {
  CatalogServiceController,
  CatalogServiceControllerMethods,
  GetProductByIdReq,
  GetProductsReq,
  PingReq,
  PingResponse,
  Product,
  ProductList,
} from '@app/proto/catalog';
import { Controller } from '@nestjs/common';
import { Observable } from 'rxjs';
import { CatalogService } from './catalog.service';

@Controller()
@CatalogServiceControllerMethods()
export class CatalogController implements CatalogServiceController {
  constructor(private readonly catalogService: CatalogService) {}
  getProducts(
    request: GetProductsReq,
  ): Promise<ProductList> | Observable<ProductList> | ProductList {
    throw new Error('Method not implemented.');
  }
  getProductById(
    request: GetProductByIdReq,
  ): Promise<Product> | Observable<Product> | Product {
    throw new Error('Method not implemented.');
  }

  ping(requst: PingReq): PingResponse {
    return this.catalogService.ping();
  }
}
