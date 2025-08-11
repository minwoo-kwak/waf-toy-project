export interface User {
  id: string;
  email: string;
  name: string;
  picture: string;
  verified: boolean;
}

export interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  loading: boolean;
}

export interface LoginResponse {
  token: string;
  user: User;
  expires_at: string;
}

export interface AuthContextType {
  authState: AuthState;
  login: (code: string, state: string) => Promise<void>;
  logout: () => void;
  getAuthUrl: () => Promise<string>;
  setAuth: (user: User, token: string) => void;
}