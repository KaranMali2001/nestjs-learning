import { CATALOG_SERVICE_NAME, CatalogServiceClient } from '@app/proto/catalog';
import { CATALOG_CLIENT } from '@app/proto/catalog.constant';
import { ClientGrpc, ClientProxy } from '@nestjs/microservices';

export const CATALOG_SVC = 'CATALOG_SVC' as const;
export const MEDIA_SVC = 'MEDIA_SVC' as const;
export const SEARCH_SVC = 'SEARCH_SVC' as const;

export const CatalogSvcProvider = {
  provide: CATALOG_SVC,
  useFactory: (client: ClientGrpc) =>
    client.getService<CatalogServiceClient>(CATALOG_SERVICE_NAME),
  inject: [CATALOG_CLIENT],
};

export const MediaSvcProvider = {
  provide: MEDIA_SVC,
  useFactory: (client: ClientProxy) => client,
  inject: ['MEDIA_SERVICE'],
};

export const SearchSvcProvider = {
  provide: SEARCH_SVC,
  useFactory: (client: ClientProxy) => client,
  inject: ['SEARCH_SERVICE'],
};
