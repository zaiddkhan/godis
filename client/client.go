package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/tidwall/resp"
	"net"
)

type Client struct {
	addr string
}

func NewClient(address string) *Client {
	return &Client{
		addr: address,
	}
}

func (c *Client) Set(ctx context.Context, key string, val string) error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)
	wr.WriteArray([]resp.Value{resp.StringValue("SET"), resp.StringValue(key), resp.StringValue(val)})
	fmt.Println("%s", buf.String())
	_, errr := conn.Write(buf.Bytes())
	return errr
}
