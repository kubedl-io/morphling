import request from "@/utils/request";

const APIV1Prefix = "/api/v1alpha1";

export async function getExperimentDetail(params) {
  return request(APIV1Prefix + `/experiment/detail`, {
    params: {
      ...params,
      page_size: 10,
      replica_type: "ALL",
      status: "ALL"
    }
  });
}
