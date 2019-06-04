/*
 * Copyright (c) 2019 Leonardo Faoro. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package cloudflare

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
