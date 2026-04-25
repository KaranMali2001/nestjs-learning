import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { ProfileMoodule } from './profile/profile.module';

@Module({
  imports: [ProfileMoodule],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
