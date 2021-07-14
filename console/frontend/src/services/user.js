import request from '@/utils/request';

export async function query() {
  return request('/api/users');
}

export async function queryConfig() {
  return request('/api/v1alpha1/data/config');
}

export async function queryCurrent() {
  return {
    accountId: '00000001',
    loginId: '00000001',
  }
  // request('/api/currentUser');
}

export async function queryNotices() {
  return request('/api/notices');
}

export async function queryNamespaces() {
  return request('/api/v1alpha1/data/namespaces');
}


export async function queryAlgorithmNames() {
  return request('/api/v1alpha1/data/algorithmNames');
}
