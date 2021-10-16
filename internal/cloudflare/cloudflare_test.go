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
		apiToken string
	}
	tests := []struct {
		name string
		args args
		want Cloudflare
	}{
		{
			"NewCloudFlareClient",
			args{
				"token",
			},
			Cloudflare{
				API:      "https://api.cloudflare.com/client/v4",
				ApiToken: "token",
				Client: http.Client{
					Timeout: time.Second * 30,
				},
			},
		},
	}
	t.Run("NewClientTokenEmpty", func(t *testing.T) {
		assert.PanicsWithValue(t, errNoAuthorization, func() { New("") })
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.apiToken); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCloudflare_Zones(t *testing.T) {

	// fail case
	cc := New("fake")
	_, err := cc.Zones()
	assert.NotNil(t, err)

	c := New(os.Getenv("CF_API_TOKEN"))
	zones, err2 := c.Zones()
	assert.Nil(t, err2)

	fmt.Println("zones", zones)

	for _, z := range zones {
		fmt.Println(z)
	}

}
