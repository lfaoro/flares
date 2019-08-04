/*
 * Copyright (c) 2019 Leonardo Faoro. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package cloudflare

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		apiKey   string
		apiEmail string
	}
	tests := []struct {
		name string
		args args
		want Cloudflare
	}{
		{
			"NewCloudFlareClient",
			args{
				"key",
				"email",
			},
			Cloudflare{
				API:       "https://api.cloudflare.com/client/v4",
				AuthKey:   "key",
				AuthEmail: "email",
				Client: http.Client{
					Timeout: time.Second * 30,
				},
			},
		},
	}
	t.Run("NewClientApiEmpty", func(t *testing.T) {
		assert.PanicsWithValue(t, "cloudflare: missing required AuthKey, AuthEmail", func() { New("", "email") })
	})
	t.Run("NewClientEmailEmpty", func(t *testing.T) {
		assert.PanicsWithValue(t, "cloudflare: missing required AuthKey, AuthEmail", func() { New("api", "") })
	})
	t.Run("NewClientWithApiAndEmailEmpty", func(t *testing.T) {
		assert.PanicsWithValue(t, "cloudflare: missing required AuthKey, AuthEmail", func() { New("", "") })
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.apiKey, tt.args.apiEmail); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCloudflare_Zones(t *testing.T) {

	// fail case
	cc := New(os.Getenv("CF_API_KEY"), "fake@email.com")
	zones, err := cc.Zones()
	assert.NotNil(t, err)

	c := New(os.Getenv("CF_API_KEY"), os.Getenv("CF_API_EMAIL"))
	zones, err = c.Zones()
	assert.Nil(t, err)

	fmt.Println("zones", zones)

	for _, z := range zones {
		fmt.Println(z)
	}

}
