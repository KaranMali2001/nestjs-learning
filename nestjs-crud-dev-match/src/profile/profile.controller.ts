import { Body, Controller, Delete, Get, HttpCode, HttpStatus, Logger, NotFoundException, Param, ParseUUIDPipe, Post, Put, Query, ValidationPipe } from '@nestjs/common';
import { CreateProfileDto } from './dto/create-profile.dto';
import { UpdateProfileDto } from './dto/update-profile.dto';
import { ProfileService } from './profile.service';

@Controller('profile')
export class ProfileController {
  constructor(private readonly profileService: ProfileService) {}
  @Get()
  getProfileByName(@Query('name') name: string) {
    if (!name) {
      Logger.debug('karan');
      Logger.warn('warn');
      Logger.error('errotr');
      Logger.debug(Logger.getTimestamp());
      throw new NotFoundException();
    }
    return this.profileService.findByName(name);
  }
  @Get(':id')
  getProfiles(@Param('id', ParseUUIDPipe) id: string) {
    return this.profileService.findAllProfile();
  }
  @Post()
  createProfile(@Body(new ValidationPipe()) body: CreateProfileDto) {
    return this.profileService.createProfile(body);
  }
  @Put(':id')
  updateProfile(@Body() body: UpdateProfileDto, @Param('id', ParseUUIDPipe) id: string) {
    return {
      updatedId: id,
      updateName: body.name,
      updateDes: body.des,
    };
  }
  @Delete(':id')
  @HttpCode(HttpStatus.NO_CONTENT)
  deleteProfile(@Param() id: string) {
    return {
      deletedId: id,
    };
  }
}
