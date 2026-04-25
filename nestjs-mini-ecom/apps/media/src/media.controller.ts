import { Controller } from '@nestjs/common';
import { MessagePattern } from '@nestjs/microservices';
import { MediaService } from './media.service';

@Controller()
export class MediaController {
  constructor(private readonly mediaService: MediaService) {}

  @MessagePattern('ping')
  ping() {
    return this.mediaService.ping();
  }
}
