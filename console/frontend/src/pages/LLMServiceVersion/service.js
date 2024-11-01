import request from '@/utils/request';

const APIV1Prefix = '/api/v1alpha1';

export async function getLLMServiceVersions() {
  return request(`${APIV1Prefix}/llm-service-version`, {
    method: 'GET',
  });
}

export async function submitLLMServiceVersion(data) {
  return request(`${APIV1Prefix}/llm-service-version`, {
    method: 'POST',
    data,
    headers: {
      'Content-Type': 'application/json',
      'X-Content-Type-Options': 'nosniff',
      'X-Frame-Options': 'DENY'
    }
  });
}
