package kache

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	pb "github.com/falldio/Kache/pkg/proto/gen"
	"google.golang.org/protobuf/proto"
)

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bts, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}
	if err = proto.Unmarshal(bts, out); err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}
	return nil
}

var _ PeerGetter = (*httpGetter)(nil)
