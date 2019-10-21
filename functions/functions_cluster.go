package functions

type ClusterFunctions struct {
	global     *GlobalClusterFunctions
	local      *LocalClusterFunctions
	localSmart *LocalSmartClusterFunctions
}

func (f *ClusterFunctions) Global() *GlobalClusterFunctions {
	return f.global
}

func (f *ClusterFunctions) Local() *LocalClusterFunctions {
	return f.local
}

func (f *ClusterFunctions) LocalSmart() *LocalSmartClusterFunctions {
	return f.localSmart
}
