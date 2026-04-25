import { Injectable } from '@nestjs/common';

@Injectable()
export class MediaService {
  ping() {
    return {
      ok: true,
      serviceName: 'MediaService',
      time: new Date().toISOString(),
    };
  }
}
