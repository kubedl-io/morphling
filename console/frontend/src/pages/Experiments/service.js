import request from "@/utils/request";

const APIV1Prefix = "/api/v1alpha1";

export async function queryPes(params) {
  const ret = await request(APIV1Prefix + "/experiment/list", {
    params
  });
  return {
    data: ret.data.peInfos,
    total: ret.data.total
  };
}

export async function deletePe(namespace, name) {
  return request(
    APIV1Prefix + `/experiment/${namespace}/${name}`,
    {
      method: "DELETE"
    }
  );
}
