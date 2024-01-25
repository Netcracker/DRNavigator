package v1

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyInto copies nested fields to given object
func (in *CR) DeepCopyInto(out *CR) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)

	if in.Spec.SiteManager.After != nil {
		out.Spec.SiteManager.After = make([]string, len(in.Spec.SiteManager.After))
		copy(out.Spec.SiteManager.After, in.Spec.SiteManager.After)
	}
	if in.Spec.SiteManager.Before != nil {
		out.Spec.SiteManager.Before = make([]string, len(in.Spec.SiteManager.Before))
		copy(out.Spec.SiteManager.Before, in.Spec.SiteManager.Before)
	}
	if in.Spec.SiteManager.Sequence != nil {
		out.Spec.SiteManager.Sequence = make([]string, len(in.Spec.SiteManager.Sequence))
		copy(out.Spec.SiteManager.Sequence, in.Spec.SiteManager.Sequence)
	}
	if in.Spec.SiteManager.AllowedStandbyStateList != nil {
		out.Spec.SiteManager.AllowedStandbyStateList = make([]string, len(in.Spec.SiteManager.AllowedStandbyStateList))
		copy(out.Spec.SiteManager.AllowedStandbyStateList, in.Spec.SiteManager.AllowedStandbyStateList)
	}
	if in.Spec.SiteManager.Timeout != nil {
		out.Spec.SiteManager.Timeout = new(int64)
		*out.Spec.SiteManager.Timeout = *in.Spec.SiteManager.Timeout
	}
}

// DeepCopyObject is an runtime.Object interface function
func (in *CR) DeepCopyObject() runtime.Object {
	out := CR{}
	in.DeepCopyInto(&out)
	return &out
}

// DeepCopyInto copies nested fields to given object
func (in *CRList) DeepCopyInto(out *CRList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)

	if in.Items != nil {
		out.Items = make([]CR, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

// DeepCopyObject is an runtime.Object interface function
func (in *CRList) DeepCopyObject() runtime.Object {
	out := CRList{}
	in.DeepCopyInto(&out)
	return &out
}
