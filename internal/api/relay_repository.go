package api

import (
	"context"
	"net/http"
	"strconv"
)

type RelayRepository struct {
	cl *Client
}

func (r RelayRepository) Test(ctx context.Context, id int) error {
	return r.cl.do(ctx, http.MethodPost, "/relays/"+strconv.Itoa(id)+"/test", nil, nil)
}

func NewRelayRepository(cl *Client) *RelayRepository {
	return &RelayRepository{cl: cl}
}
