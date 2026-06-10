package v1

// =================================================
// Server implementation
// =================================================

type RPCServer struct {
	Impl Client
}

// =================================================
// OCM Cluster Operations
// =================================================

func (s *RPCServer) GetCluster(args *GetClusterArgs, reply *GetClusterReply) error {
	cluster, err := s.Impl.GetCluster(args.Key)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	clusterJSON, err := SerializeCluster(cluster)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.ClusterJSON = clusterJSON
	return nil
}

func (s *RPCServer) GetClusterAnyStatus(args *GetClusterAnyStatusArgs, reply *GetClusterAnyStatusReply) error {
	cluster, err := s.Impl.GetClusterAnyStatus(args.ClusterID)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	clusterJSON, err := SerializeCluster(cluster)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.ClusterJSON = clusterJSON
	return nil
}

func (s *RPCServer) GetClusters(args *GetClustersArgs, reply *GetClustersReply) error {
	clusters, err := s.Impl.GetClusters(args.ClusterIDs)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	clustersJSON, err := SerializeClusters(clusters)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.ClustersJSON = clustersJSON
	return nil
}

func (s *RPCServer) GetManagementCluster(args *GetManagementClusterArgs, reply *GetManagementClusterReply) error {
	mgmtCluster, err := s.Impl.GetManagementCluster(args.ClusterID)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	mgmtClusterJSON, err := SerializeCluster(mgmtCluster)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.ManagementClusterJSON = mgmtClusterJSON
	return nil
}

// =================================================
// OCM Subscription & Organization
// =================================================

func (s *RPCServer) GetSubscription(args *GetSubscriptionArgs, reply *GetSubscriptionReply) error {
	subscription, err := s.Impl.GetSubscription(args.Key)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	subscriptionJSON, err := SerializeSubscription(subscription)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.SubscriptionJSON = subscriptionJSON
	return nil
}

func (s *RPCServer) GetOrganization(args *GetOrganizationArgs, reply *GetOrganizationReply) error {
	organization, err := s.Impl.GetOrganization(args.OrgID)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	organizationJSON, err := SerializeOrganization(organization)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.OrganizationJSON = organizationJSON
	return nil
}

// =================================================
// AWS Account Operations
// =================================================

func (s *RPCServer) GetSupportRoleArnForCluster(args *GetSupportRoleArnForClusterArgs, reply *GetSupportRoleArnForClusterReply) error {
	arn, err := s.Impl.GetSupportRoleArnForCluster(args.ClusterID)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.Arn = arn
	return nil
}

func (s *RPCServer) GetAWSAccountIdForCluster(args *GetAWSAccountIdForClusterArgs, reply *GetAWSAccountIdForClusterReply) error {
	accountID, err := s.Impl.GetAWSAccountIdForCluster(args.ClusterID)
	if err != nil {
		reply.Error = err.Error()
		return nil
	}

	reply.AccountID = accountID
	return nil
}
