// +build !ignore_autogenerated

// Code generated by main. DO NOT EDIT.

package v1alpha1

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionSelection) DeepCopyInto(out *VersionSelection) {
	*out = *in
	if in.Semver != nil {
		in, out := &in.Semver, &out.Semver
		*out = new(VersionSelectionSemver)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionSelection.
func (in *VersionSelection) DeepCopy() *VersionSelection {
	if in == nil {
		return nil
	}
	out := new(VersionSelection)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionSelectionSemver) DeepCopyInto(out *VersionSelectionSemver) {
	*out = *in
	if in.Prereleases != nil {
		in, out := &in.Prereleases, &out.Prereleases
		*out = new(VersionSelectionSemverPrereleases)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionSelectionSemver.
func (in *VersionSelectionSemver) DeepCopy() *VersionSelectionSemver {
	if in == nil {
		return nil
	}
	out := new(VersionSelectionSemver)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VersionSelectionSemverPrereleases) DeepCopyInto(out *VersionSelectionSemverPrereleases) {
	*out = *in
	if in.Identifiers != nil {
		in, out := &in.Identifiers, &out.Identifiers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VersionSelectionSemverPrereleases.
func (in *VersionSelectionSemverPrereleases) DeepCopy() *VersionSelectionSemverPrereleases {
	if in == nil {
		return nil
	}
	out := new(VersionSelectionSemverPrereleases)
	in.DeepCopyInto(out)
	return out
}
