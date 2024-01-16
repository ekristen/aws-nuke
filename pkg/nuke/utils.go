package nuke

import (
	"github.com/ekristen/libnuke/pkg/types"
)

func ResolveResourceTypes(
	base types.Collection,
	mapping map[string]string,
	include, exclude, cloudControl []types.Collection) types.Collection {
	for _, cl := range cloudControl {
		oldStyle := types.Collection{}
		for _, c := range cl {
			os, found := mapping[c]
			if found {
				oldStyle = append(oldStyle, os)
			}
		}

		base = base.Union(cl)
		base = base.Remove(oldStyle)
	}

	for _, i := range include {
		if len(i) > 0 {
			base = base.Intersect(i)
		}
	}

	for _, e := range exclude {
		base = base.Remove(e)
	}

	return base
}
