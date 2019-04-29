package cloud

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudflare_AllZones(t *testing.T) {

	c := NewCloudflare("", "") // make sure you have the envvars set.
	zones, err := c.AllZones()
	assert.Nil(t, err)

	fmt.Println("zones", zones)

	for _, z := range zones {
		fmt.Println(z)
	}

}
