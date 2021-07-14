# How to DEBUG

## DEBUG with local process 

- Credentials

    To run Morphling locally, you must have the access to the kubernetes cluster, the credential is a distributed
    kube-config cert file.

- Install CRDs and run Morphling components
    
    ```bash
    # kube-config cert file
    export KUBECONFIG=${PATH_TO_CONFIG}
   
    # install CRDs
    make install
  
    # install basic components, e.g., PV, mysql-db
    kubectl apply -f manifests/configmap
    kubectl apply -f manifests/pv
    kubectl apply -f manifests/mysql-db
    kubectl apply -f manifests/db-manager
  
    # run Morphling controller locally
    make run
    ```

## Troubleshoot with Pod

The followings are the steps to Troubleshoot Morphling using Pod.

- Check if all components are running successfully:
`kubectl get deployment -n morphling-system`. If succeed, you will see
three _**ready**_ deployments: morphling-controller, morphling-db-manager, and morphling-mysql.

- Check the Morphling controller logs manually `kubectl -n morphling-system logs morphling-controller-XXX ` for debugging.

- Check your cluster has enabled Kubernetes DNS service by `kubectl get svc -n kube-system kube-dns`. See this [GitHub Issue](https://github.com/mattermost/mattermost-docker/issues/419) for detailed discussion.

- Check the ClusterRoleBinding to make sure morphling-controller has been granted corresponding authorities by `kubectl get ClusterRoleBinding morphling-controller`. If you cannot grant Morphling the cluster-scope authorities, you may need
to change the ClusterRole to Role in the [manifest yaml](../manifests/controllers/rbac.yaml). 
See this [GitHub Issue](https://github.com/kubernetes-sigs/kubebuilder/issues/1366) for detailed discussion.

