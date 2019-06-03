/*
 * Copyright (c) 2019 Leonardo Faoro. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package cloud

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudflare_AllZones(t *testing.T) {

	c := NewCloudflare("", "") // make sure you have the envvars set.
	zones, err := c.Zones()
	assert.Nil(t, err)

	fmt.Println("zones", zones)

	for _, z := range zones {
		fmt.Println(z)
	}

}
