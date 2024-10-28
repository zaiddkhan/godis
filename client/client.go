package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/tidwall/resp"
	"log"
	"net"
)

type Client struct {
	addr string
	conn net.Conn
}

func NewClient(address string) *Client {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	return &Client{
		addr: address,
		conn: conn,
	}
}

func (c *Client) Set(ctx context.Context, key string, val string) error {
	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)
	wr.WriteArray([]resp.Value{resp.StringValue("SET"), resp.StringValue(key), resp.StringValue(val)})
	fmt.Println("%s", buf.String())
	_, err := c.conn.Write(buf.Bytes())
	buf.Reset()
	return err
}

func (c *Client) Get(ctx context.Context, key string) error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	wr := resp.NewWriter(buf)
	wr.WriteArray([]resp.Value{resp.StringValue("GET"), resp.StringValue(key)})
	fmt.Println("%s", buf.String())
	_, err = conn.Write(buf.Bytes())
	buf.Reset()
	return err
}
