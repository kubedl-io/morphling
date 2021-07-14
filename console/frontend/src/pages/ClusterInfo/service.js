import request from "@/utils/request";

const APIV1Prefix = "/api/v1alpha1";

export async function getOverviewTotal() {
  return request(APIV1Prefix + "/data/total");
}

export async function getOverviewRequestPodPhase() {
  return request(APIV1Prefix + `/data/request/Running`);
}

export async function getOverviewNodeInfos(params) {
  const ret = await request(APIV1Prefix + "/data/nodeInfos", {
    params
  });
  return ret;
}
