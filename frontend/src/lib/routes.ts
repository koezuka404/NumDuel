import type { UserRole } from '../types/dto';

export function homePathForRole(role: UserRole): '/admin' | '/matching' {
  return role === 'master' ? '/admin' : '/matching';
}
