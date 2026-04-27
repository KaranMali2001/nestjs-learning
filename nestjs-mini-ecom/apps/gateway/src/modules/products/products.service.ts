import type { CatalogServiceClient } from '@app/proto/catalog';
import { Inject, Injectable } from '@nestjs/common';
import { ClientProxy } from '@nestjs/microservices';
import { firstValueFrom } from 'rxjs';
import { CATALOG_SVC, MEDIA_SVC } from '../../shared/clients.provider';

@Injectable()
export class ProductService {
  constructor(
    @Inject(CATALOG_SVC) private readonly catalogSvc: CatalogServiceClient,
    @Inject(MEDIA_SVC) private readonly mediaSvc: ClientProxy,
  ) {}
  async pingCatalogAndMedia() {
    return await Promise.all([firstValueFrom(this.catalogSvc.ping({})), firstValueFrom(this.mediaSvc.send('ping', {}))]);
  }
}
