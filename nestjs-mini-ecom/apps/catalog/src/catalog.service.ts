import { PingResponse } from '@app/proto/catalog';
import { CATALOG_SERVICE } from '@app/proto/catalog.constant';
import { Injectable } from '@nestjs/common';

@Injectable()
export class CatalogService {
  ping(): PingResponse {
    return {
      ok: true,
      serviceName: CATALOG_SERVICE,
      time: new Date().toISOString(),
    };
  }
}
