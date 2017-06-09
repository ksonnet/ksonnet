package utils

import (
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

// DependencyOrder is a `sort.Interface` that *best-effort* sorts the
// objects so that known dependencies appear earlier in the list.  The
// idea is to prevent *some* of the "crash-restart" loops when
// creating inter-dependent resources.
type DependencyOrder []*runtime.Unstructured

func (l DependencyOrder) Len() int      { return len(l) }
func (l DependencyOrder) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l DependencyOrder) Less(i, j int) bool {
	return depTier(l[i]) < depTier(l[j])
}

// AlphabeticalOrder is a `sort.Interface` that sorts the
// objects by namespace/name/kind alphabetical order
type AlphabeticalOrder []*runtime.Unstructured

func (l AlphabeticalOrder) Len() int      { return len(l) }
func (l AlphabeticalOrder) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l AlphabeticalOrder) Less(i, j int) bool {
	a, b := l[i], l[j]

	if a.GetNamespace() != b.GetNamespace() {
		return a.GetNamespace() < b.GetNamespace()
	}
	if a.GetName() != b.GetName() {
		return a.GetName() < b.GetName()
	}
	return a.GetKind() < b.GetKind()
}
