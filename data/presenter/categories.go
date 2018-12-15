package presenter

import (
	"context"
	"net/http"
	"sort"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
)

type Node struct {
	*data.Category
	Values []*Node `json:"values"`

	// parent node
	parent *Node
}

type sortedCategory []*Node

func (a sortedCategory) Len() int           { return len(a) }
func (a sortedCategory) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortedCategory) Less(i, j int) bool { return a[i].ID < a[j].ID }

func (n *Node) Render(w http.ResponseWriter, r *http.Request) error {
	if n.Values == nil {
		n.Values = make([]*Node, 0)
	}
	for _, v := range n.Values {
		v.Render(w, r)
	}
	return nil
}

func NewCategory(ctx context.Context, category *data.Category) *Node {
	c := &Node{
		Category: category,
	}
	descendents, err := data.DB.Category.FindDescendants(category.ID)
	if err != nil {
		return c
	}
	c.Values = newCategoryList(descendents)
	return c
}

// organize a list of categories into a category tree
func newCategoryList(categories []*data.Category) []*Node {
	list := []*Node{}

	cache := map[int64]*Node{}
	for _, c := range categories {
		cache[c.ID] = &Node{Category: c}
	}

	for _, c := range categories {
		// if there's an key in the cache that's >less< than the current
		// id and that their keyed left and right values covers the id.
		//
		// ie-> id is BETWEEN cache[key].Left AND cache[key].Right then
		// it is a >child< of said cache[key]
		var parent *Node
		for _, v := range cache {
			// check if 'v' could be a parent (c.ID is between v.Left and [v.Right])
			if v.Left < c.ID && v.Right >= c.ID {
				if parent == nil {
					// starting value is the first valid value
					parent = v
				} else if parent.Left < v.ID && parent.Right >= v.ID {
					// check if this is the nearest parent found
					parent = v
				}
			}
		}
		if parent != nil {
			// if the code is not a root node, append it to the nearest
			// parent values, and set its parent to such
			parent.Values = append(parent.Values, cache[c.ID])
			cache[c.ID].parent = parent
		}
	}

	for _, v := range cache {
		// throw away the leaf nodes because they should be contained under
		// the parent values
		if v.parent == nil {
			list = append(list, v)
		}
	}
	sort.Sort(sortedCategory(list))

	return list
}

func NewCategoryList(ctx context.Context, categories []*data.Category) []render.Renderer {
	list := newCategoryList(categories)

	presented := make([]render.Renderer, len(list))
	for i, v := range list {
		presented[i] = v
	}

	return presented
}
