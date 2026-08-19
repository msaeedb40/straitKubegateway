import { Injectable, signal, computed } from '@angular/core';
import { UserProfile, AuthTokens, LoginCredentials } from './auth.types';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly userSignal = signal<UserProfile | null>({
    id: 'usr-admin-01',
    username: 'cluster-admin',
    email: 'admin@straitkube.io',
    roles: ['ClusterAdmin', 'NetworkOperator'],
    permissions: ['*'],
    tenantId: 'system'
  });

  private readonly tokensSignal = signal<AuthTokens | null>({
    accessToken: 'strait_jwt_mock_token_production',
    expiresAt: Date.now() + 86400000
  });

  readonly user = this.userSignal.asReadonly();
  readonly tokens = this.tokensSignal.asReadonly();
  readonly isAuthenticated = computed(() => !!this.userSignal() && !!this.tokensSignal());
  readonly userRoles = computed(() => this.userSignal()?.roles ?? []);

  login(credentials: LoginCredentials): Promise<boolean> {
    this.userSignal.set({
      id: 'usr-admin-01',
      username: credentials.username || 'cluster-admin',
      email: `${credentials.username || 'admin'}@straitkube.io`,
      roles: ['ClusterAdmin', 'NetworkOperator'],
      permissions: ['*'],
      tenantId: 'system'
    });
    this.tokensSignal.set({
      accessToken: 'strait_jwt_mock_token_production',
      expiresAt: Date.now() + 86400000
    });
    return Promise.resolve(true);
  }

  logout(): void {
    this.userSignal.set(null);
    this.tokensSignal.set(null);
  }

  hasRole(role: string): boolean {
    return this.userRoles().includes(role) || this.userRoles().includes('ClusterAdmin');
  }

  hasPermission(permission: string): boolean {
    const permissions = this.userSignal()?.permissions ?? [];
    return permissions.includes('*') || permissions.includes(permission);
  }
}
