// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

func (s *SubHandler) Listen() {

	go func() {
		exitCh := make(chan bool)
		_ = s.subService.Subscribe(exitCh)
	}()
}
