import { ClerkClient, createClerkClient, getAuth } from '@clerk/express';
import { CanActivate, ExecutionContext, Injectable, UnauthorizedException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Reflector } from '@nestjs/core';
import { Request } from 'express';
import { Env } from '../config/env';

@Injectable()
export class ClerkAuthGuard implements CanActivate {
  private clerkClient: ClerkClient;

  constructor(
    private config: ConfigService<Env, true>,
    private reflector: Reflector,
  ) {
    this.clerkClient = createClerkClient({
      secretKey: this.config.getOrThrow('CLERK_SECRET_KEY'),
    });
  }
  async canActivate(context: ExecutionContext): Promise<boolean> {
    const isPublic = this.reflector.get<boolean>('isPublic', context.getHandler());
    if (isPublic) return true;

    const req = context.switchToHttp().getRequest<Request>();
    const { isAuthenticated, userId } = getAuth(req);

    if (!isAuthenticated || !userId) {
      throw new UnauthorizedException('Invalid Token or No Token');
    }

    const user = await this.clerkClient.users.getUser(userId);

    req.user = {
      id: user.id,
      email: user.emailAddresses[0]?.emailAddress ?? '',
      role: (user.publicMetadata?.role as string) ?? 'user',
    };

    return true;
  }
}
