package utils

import (
	"sort"

	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/runtime"
)

var (
	gkNamespace    = unversioned.GroupKind{Group: "", Kind: "Namespace"}
	gkTpr          = unversioned.GroupKind{Group: "extensions", Kind: "ThirdPartyResource"}
	gkStorageClass = unversioned.GroupKind{Group: "storage.k8s.io", Kind: "StorageClass"}

	gkPod         = unversioned.GroupKind{Group: "", Kind: "Pod"}
	gkJob         = unversioned.GroupKind{Group: "batch", Kind: "Job"}
	gkDeployment  = unversioned.GroupKind{Group: "extensions", Kind: "Deployment"}
	gkDaemonSet   = unversioned.GroupKind{Group: "extensions", Kind: "DaemonSet"}
	gkStatefulSet = unversioned.GroupKind{Group: "apps", Kind: "StatefulSet"}
)

// These kinds all start pods.
// TODO: expand this list.
func isPodOrSimilar(gk unversioned.GroupKind) bool {
	return gk == gkPod ||
		gk == gkJob ||
		gk == gkDeployment ||
		gk == gkDaemonSet ||
		gk == gkStatefulSet
}

// Arbitrary numbers used to do a simple topological sort of resources.
// TODO: expand this list.
func depTier(o *runtime.Unstructured) int {
	gk := o.GroupVersionKind().GroupKind()
	if gk == gkNamespace || gk == gkTpr || gk == gkStorageClass {
		return 10
	} else if isPodOrSimilar(gk) {
		return 100
	} else {
		return 50
	}
}

type dependentObjects []*runtime.Unstructured

func (l dependentObjects) Len() int      { return len(l) }
func (l dependentObjects) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l dependentObjects) Less(i, j int) bool {
	return depTier(l[i]) < depTier(l[j])
}

// SortDepFirst *best-effort* sorts the objects so that known
// dependencies appear earlier in the list.  The idea is to prevent
// *some* of the "crash-restart" loops when creating inter-dependent
// resources.
func SortDepFirst(objs []*runtime.Unstructured) {
	sort.Sort(dependentObjects(objs))
}
