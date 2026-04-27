import { CanActivate, ExecutionContext, ForbiddenException, Injectable } from '@nestjs/common';
import { Reflector } from '@nestjs/core';
import { Request } from 'express';

@Injectable()
export class RoleGuard implements CanActivate {
  constructor(private reflector: Reflector) {}
  canActivate(context: ExecutionContext): boolean {
    const requiredRoles = this.reflector.get<string[]>('roles', context.getHandler());
    if (!requiredRoles?.length) {
      return true;
    }
    const { user } = context.switchToHttp().getRequest<Request>();
    if (!requiredRoles.includes(user?.role ?? '')) {
      throw new ForbiddenException('Insufficient permissions');
    }
    return true;
  }
}
