import client from './client'

export const authApi = {
  login: (token: string) =>
    client.post('/auth/login', { token }),

  logout: () =>
    client.post('/auth/logout'),
}
