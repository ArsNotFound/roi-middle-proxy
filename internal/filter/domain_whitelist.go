package filter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/infobloxopen/go-trees/domain"
	"github.com/infobloxopen/go-trees/domaintree"
)

func DomainInList(list []string) (goproxy.ReqConditionFunc, error) {
	tree := &domaintree.Node{}
	for _, host := range list {
		d, err := domain.MakeNameFromString(host)
		if err != nil {
			return nil, fmt.Errorf("failed to make domain label: %w", err)
		}
		tree = tree.Insert(d, struct{}{})
	}

	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
		host := strings.Split(req.Host, ":")[0]

		d, err := domain.MakeNameFromString(host)
		if err != nil {
			return false
		}

		_, present := tree.Get(d)
		return present
	}, nil
}
