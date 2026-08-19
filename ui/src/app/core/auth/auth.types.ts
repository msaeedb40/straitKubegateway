export interface UserProfile {
  id: string;
  username: string;
  email: string;
  roles: string[];
  permissions: string[];
  tenantId?: string;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken?: string;
  expiresAt: number;
}

export interface AuthState {
  user: UserProfile | null;
  tokens: AuthTokens | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface LoginCredentials {
  username: string;
  password?: string;
  token?: string;
}
