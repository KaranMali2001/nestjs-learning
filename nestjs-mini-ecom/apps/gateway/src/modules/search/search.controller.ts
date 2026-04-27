import { Controller, Get, Inject } from '@nestjs/common';
import type { ClientProxy } from '@nestjs/microservices';
import { firstValueFrom } from 'rxjs';
import { Public } from '../../decorators/public.decorator';
import { SEARCH_SVC } from '../../shared/clients.provider';

@Controller('search')
export class SearchController {
  constructor(@Inject(SEARCH_SVC) private readonly searchSvc: ClientProxy) {}
  @Public()
  @Get('ping')
  async pingSearch() {
    return firstValueFrom(this.searchSvc.send('ping', {}));
  }
}
