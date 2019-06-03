/*
 * Copyright (c) 2019 Leonardo Faoro. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package cloud

// Service is a dns cloud provider.
type Service interface {
	TableFor(string) ([]byte, error)
}
