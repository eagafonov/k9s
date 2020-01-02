package client

import (
	"fmt"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"vbom.ml/util/sortorder"
)

// GVR represents a kubernetes resource schema as a string.
// Format is group/version/resources
type GVR struct {
	raw, g, v, r, sr string
}

// NewGVR builds a new gvr from a group, version, resource.
func NewGVR(gvr string) GVR {
	var g, v, r, sr string

	tokens := strings.Split(gvr, ":")
	raw := gvr
	if len(tokens) == 2 {
		raw, sr = tokens[0], tokens[1]
	}
	tokens = strings.Split(raw, "/")
	switch len(tokens) {
	case 3:
		g, v, r = tokens[0], tokens[1], tokens[2]
	case 2:
		v, r = tokens[0], tokens[1]
	case 1:
		r = tokens[0]
	default:
		panic(fmt.Sprintf("can't parse GVR %q", gvr))
	}

	return GVR{raw: gvr, g: g, v: v, r: r, sr: sr}
}

// NewGVRFromMeta builds a gvr from resource metadata.
func NewGVRFromMeta(a metav1.APIResource) GVR {
	return GVR{
		raw: path.Join(a.Group, a.Version, a.Name),
		g:   a.Group,
		v:   a.Version,
		r:   a.Name,
	}
}

// FromGVAndR builds a gvr from a group/version and resource.
func FromGVAndR(gv, r string) GVR {
	return NewGVR(path.Join(gv, r))
}

// AsResourceName returns a resource . separated descriptor in the shape of kind.version.group.
func (g GVR) AsResourceName() string {
	return g.r + "." + g.v + "." + g.g
}

// SubResource returns a sub resource if available.
func (g GVR) SubResource() string {
	return g.sr
}

// String returns gvr as string.
func (g GVR) String() string {
	return g.raw
}

// AsGV returns the group version scheme representation.
func (g GVR) AsGV() schema.GroupVersion {
	return schema.GroupVersion{
		Group:   g.g,
		Version: g.v,
	}
}

// AsGVR returns a a full schema representation.
func (g GVR) AsGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    g.ToG(),
		Version:  g.ToV(),
		Resource: g.ToR(),
	}
}

// ToV returns the resource version.
func (g GVR) ToV() string {
	return g.v
}

// ToRAndG returns the resource and group.
func (g GVR) ToRAndG() (string, string) {
	return g.r, g.g
}

// ToR returns the resource name.
func (g GVR) ToR() string {
	return g.r
}

// ToG returns the resource group name.
func (g GVR) ToG() string {
	return g.g
}

// GVRs represents a collection of gvr.
type GVRs []GVR

// Len returns the list size.
func (g GVRs) Len() int {
	return len(g)
}

// Swap swaps list values.
func (g GVRs) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

// Less returns true if i < j.
func (g GVRs) Less(i, j int) bool {
	g1, g2 := g[i].ToG(), g[j].ToG()

	return sortorder.NaturalLess(g1, g2)
}

// Helper...

// Can determines the available actions for a given resource.
func Can(verbs []string, v string) bool {
	for _, verb := range verbs {
		candidates, err := mapVerb(v)
		if err != nil {
			log.Error().Err(err).Msgf("verb mapping failed")
			return false
		}
		for _, c := range candidates {
			if verb == c {
				return true
			}
		}
	}

	return false
}

func mapVerb(v string) ([]string, error) {
	switch v {
	case "describe":
		return []string{"get"}, nil
	case "view":
		return []string{"get", "list"}, nil
	case "delete":
		return []string{"delete"}, nil
	case "edit":
		return []string{"patch", "update"}, nil
	default:
		return []string{}, fmt.Errorf("no standard verb for %q", v)
	}
}