// Package httpjson is the REFERENCE transport for sohocloud-protocol. It speaks
// the Coordinator interface over HTTP+JSON.
//
// It lives in a subpackage that the protocol core never imports: HTTP+JSON is
// the transport WE run, not the wire itself. A conformant implementation is
// free to speak the same messages over any transport — gRPC, a queue, anything
// — by implementing coordinator.Coordinator against the canonical signing bytes
// documented in SPEC.md. The JSON used here is a transport encoding only;
// signatures are computed over canon bytes, never over this JSON, so a
// different transport interoperates without renegotiating signatures.
package httpjson

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/NTARI-RAND/sohocloud-protocol/coordinator"
	"github.com/NTARI-RAND/sohocloud-protocol/employment"
	"github.com/NTARI-RAND/sohocloud-protocol/fees"
	"github.com/NTARI-RAND/sohocloud-protocol/identity"
	"github.com/NTARI-RAND/sohocloud-protocol/listing"
	"github.com/NTARI-RAND/sohocloud-protocol/liveness"
)

// Client is a coordinator.Coordinator that talks to a coordinator's HTTP+JSON
// endpoint. BaseURL is the coordinator root (no trailing slash). HTTP is
// optional; http.DefaultClient is used when nil.
type Client struct {
	BaseURL string
	HTTP    *http.Client
	// MaxResponseBytes bounds successful bodies; zero leaves size unbounded.
	MaxResponseBytes int64
}

// ErrResponseTooLarge identifies a successful body exceeding the local byte budget.
var ErrResponseTooLarge = errors.New("sohocloud: successful response exceeds byte budget")

// compile-time assertion that Client satisfies the interface.
var _ coordinator.Coordinator = (*Client)(nil)

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

func (c *Client) postJSON(ctx context.Context, path string, in any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, nil)
}

func (c *Client) getJSON(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	if c.MaxResponseBytes < 0 {
		return fmt.Errorf("sohocloud: MaxResponseBytes must be non-negative")
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("sohocloud: %s %s: status %d: %s", req.Method, req.URL.Path, resp.StatusCode, string(msg))
	}
	return decodeSuccess(resp, out, c.MaxResponseBytes)
}

func decodeSuccess(resp *http.Response, out any, limit int64) error {
	var reader io.Reader = resp.Body
	if limit > 0 {
		data, err := readSuccessBody(resp, limit)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	if out == nil {
		return nil
	}
	return decodeDocument(reader, out)
}

func decodeDocument(reader io.Reader, out any) error {
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(out); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("sohocloud: trailing response JSON")
	}
	return nil
}

func readSuccessBody(resp *http.Response, limit int64) ([]byte, error) {
	if resp.ContentLength > limit {
		return nil, ErrResponseTooLarge
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return nil, err
	}
	var extra [1]byte
	n, err := io.ReadFull(resp.Body, extra[:])
	if n > 0 {
		return nil, ErrResponseTooLarge
	}
	if err != io.EOF {
		return nil, err
	}
	return data, nil
}

// SubmitListing posts a signed capability listing.
func (c *Client) SubmitListing(ctx context.Context, l listing.CapabilityListing) error {
	return c.postJSON(ctx, "/v0/listing", l)
}

// Heartbeat posts a signed liveness signal.
func (c *Client) Heartbeat(ctx context.Context, h liveness.Heartbeat) error {
	return c.postJSON(ctx, "/v0/heartbeat", h)
}

// PollJobs fetches assignments offered to the given node.
func (c *Client) PollJobs(ctx context.Context, id identity.NodeID) ([]employment.Assignment, error) {
	var out []employment.Assignment
	err := c.getJSON(ctx, "/v0/jobs?node_id="+url.QueryEscape(string(id)), &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Decline posts a signed refusal of an assignment.
func (c *Client) Decline(ctx context.Context, d employment.Decline) error {
	return c.postJSON(ctx, "/v0/decline", d)
}

// ReportJob posts a signed job outcome.
func (c *Client) ReportJob(ctx context.Context, r employment.JobReport) error {
	return c.postJSON(ctx, "/v0/report", r)
}

// Fees fetches the coordinator's current signed fee declaration.
func (c *Client) Fees(ctx context.Context) (fees.FeeDeclaration, error) {
	var out fees.FeeDeclaration
	err := c.getJSON(ctx, "/v0/fees", &out)
	if err != nil {
		return fees.FeeDeclaration{}, err
	}
	return out, nil
}
