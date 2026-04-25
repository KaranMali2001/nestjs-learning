import { Injectable } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { CreateProfileDto } from './dto/create-profile.dto';

@Injectable()
export class ProfileService {
  private profiles = [
    {
      id: randomUUID(),
      name: 'Karan',
      des: ' des',
    },
    {
      id: randomUUID(),
      name: 'Karan Mali 2',
      des: ' des 2',
    },
  ];
  findAllProfile() {
    return this.profiles;
  }
  findByName(name: string) {
    return this.profiles.find((n) => n.name === name);
  }
  createProfile(body: CreateProfileDto) {
    this.profiles.push({
      id: randomUUID(),
      name: body.name,
      des: body.des,
    });
    return this.profiles;
  }
}
