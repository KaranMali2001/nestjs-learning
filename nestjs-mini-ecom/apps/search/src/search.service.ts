import { Injectable } from '@nestjs/common';

@Injectable()
export class SearchService {
  ping() {
    return {
      ok: true,
      serviceName: 'Search Service',
      time: new Date().toISOString(),
    };
  }
}
