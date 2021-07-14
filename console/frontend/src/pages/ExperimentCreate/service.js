import request from "@/utils/request";

const APIV1Prefix = "/api/v1alpha1";

export async function submitPeYaml(data) {
  return request(APIV1Prefix + `/experiment/submitYaml`, {
    method: "POST",
    params: {},
    data: data
  });
}

export async function submitPePars(data) {
  return request(APIV1Prefix + `/experiment/submitPars`, {
    method: "POST",
    params: {},
    data: data
  });
}
