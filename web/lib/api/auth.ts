import { request, setTokens, clearTokens } from './client';

export interface SignInResponse {
  access_token: string;
  refresh_token: string;
}

export async function signIn(username: string, password: string): Promise<SignInResponse> {
  const res = await request<{ code: number; data: SignInResponse }>('/auth/signin', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
  setTokens(res.data.access_token, res.data.refresh_token);
  return res.data;
}

export async function signOut(): Promise<void> {
  try {
    await request('/auth/signout', { method: 'POST' });
  } finally {
    clearTokens();
  }
}

export async function changePassword(oldPassword: string, newPassword: string): Promise<void> {
  await request('/auth/password', {
    method: 'POST',
    body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
  });
}
