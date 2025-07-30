package vault

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/httpclient"
)

type Vault struct {
	Client                *httpclient.Client
	Address, Token, Mount string
}

func (v *Vault) path(ns, key string, metadata bool) (string, error) {
	if ns == "" || key == "" || strings.Contains(key, "..") || strings.ContainsAny(key, "?#\\") {
		return "", fmt.Errorf("%w: invalid secret key", entity.ErrInvalid)
	}
	part := "data"
	if metadata {
		part = "metadata"
	}
	return strings.TrimRight(v.Address, "/") + "/v1/" + url.PathEscape(v.Mount) + "/" + part + "/" + url.PathEscape(ns) + "/" + url.PathEscape(key), nil
}
func (v *Vault) Get(ctx context.Context, ns, key string) (string, error) {
	path, err := v.path(ns, key, false)
	if err != nil {
		return "", err
	}
	var result struct {
		Data struct {
			Data struct {
				Value string `json:"value"`
			} `json:"data"`
		} `json:"data"`
	}
	err = v.Client.Do(ctx, "GET", path, nil, map[string]string{"X-Vault-Token": v.Token}, &result)
	return result.Data.Data.Value, err
}
func (v *Vault) Put(ctx context.Context, ns, key, value string) error {
	path, err := v.path(ns, key, false)
	if err != nil {
		return err
	}
	return v.Client.Do(ctx, "POST", path, map[string]any{"data": map[string]string{"value": value}}, map[string]string{"X-Vault-Token": v.Token}, nil)
}
func (v *Vault) Delete(ctx context.Context, ns, key string) error {
	path, err := v.path(ns, key, true)
	if err != nil {
		return err
	}
	return v.Client.Do(ctx, "DELETE", path, nil, map[string]string{"X-Vault-Token": v.Token}, nil)
}
