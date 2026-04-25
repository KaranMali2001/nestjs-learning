import { Controller, Get } from '@nestjs/common';
import { firstValueFrom } from 'rxjs';
import { GatewayService } from './gateway.service';

@Controller()
export class GatewayController {
  constructor(private readonly gatewayService: GatewayService) {}

  @Get('ping')
  async ping() {
    const [catalog, media] = await Promise.all([
      firstValueFrom(this.gatewayService.pingCatalog()),
      firstValueFrom(this.gatewayService.pingMedia()),
    ]);
    return {
      catalog,
      media,
    };
  }
}
