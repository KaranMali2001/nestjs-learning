import { CATALOG_SERVICE_NAME, CatalogServiceClient } from '@app/proto/catalog';
import { CATALOG_CLIENT } from '@app/proto/catalog.constant';
import { Inject, Injectable, OnModuleInit } from '@nestjs/common';
import type { ClientGrpc, ClientProxy } from '@nestjs/microservices';

@Injectable()
export class GatewayService implements OnModuleInit {
  private catalogSvc: CatalogServiceClient;
  private mediaSvc: ClientProxy;
  constructor(
    @Inject(CATALOG_CLIENT) private readonly catalogClient: ClientGrpc,
    @Inject('MEDIA_SERVICE') private readonly mediaClient: ClientProxy,
  ) {}
  onModuleInit() {
    this.catalogSvc =
      this.catalogClient.getService<CatalogServiceClient>(CATALOG_SERVICE_NAME);
    this.mediaSvc = this.mediaClient;
  }
  pingCatalog() {
    return this.catalogSvc.ping({});
  }
  pingMedia() {
    return this.mediaSvc.send('ping', {});
  }
}
