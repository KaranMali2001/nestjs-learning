import { Module } from '@nestjs/common';
import { SearchClientModule } from '../../shared/search-client.module';
import { SearchController } from './search.controller';

@Module({
  controllers: [SearchController],
  imports: [SearchClientModule],
})
export class SearchModule {}
