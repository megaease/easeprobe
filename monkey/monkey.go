/*
 * Copyright (c) 2022, MegaEase
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package monkey is a library to patch functions and methods at runtime for testing
package monkey

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/agiledragon/gomonkey/v2"
)

var (
	patchesMap sync.Map
)

// Patch replaces a function with another
func Patch(target, replacement interface{}) *gomonkey.Patches {
	key := fmt.Sprintf("%v", target)

	// If an existing patch exists, reset and delete it
	if existingPatches, ok := patchesMap.Load(key); ok {
		existingPatches.(*gomonkey.Patches).Reset()
		patchesMap.Delete(key)
	}

	// Apply the new patch
	patches := gomonkey.ApplyFunc(target, replacement)
	patchesMap.Store(key, patches)
	wait()
	return patches
}

// Unpatch removes a patch
func Unpatch(target interface{}) bool {
	key := fmt.Sprintf("%v", target)

	patches, ok := patchesMap.Load(key)
	if !ok {
		return false
	}
	patches.(*gomonkey.Patches).Reset()
	patchesMap.Delete(key)
	wait()
	return true
}

// PatchInstanceMethod replaces an instance method methodName for the type target with replacement
func PatchInstanceMethod(target reflect.Type, methodName string, replacement interface{}) *gomonkey.Patches {
	key := fmt.Sprintf("%v:%v", target, methodName)

	// If an existing patch exists, reset and delete it
	if existingPatches, ok := patchesMap.Load(key); ok {
		existingPatches.(*gomonkey.Patches).Reset()
		patchesMap.Delete(key)
	}

	// Apply the new patch
	patches := gomonkey.ApplyMethod(target, methodName, replacement)
	patchesMap.Store(key, patches)
	wait()
	return patches
}

// UnpatchInstanceMethod removes a patch from an instance method
func UnpatchInstanceMethod(target reflect.Type, methodName string) bool {
	key := fmt.Sprintf("%v:%v", target, methodName)

	patches, ok := patchesMap.Load(key)
	if !ok {
		return false
	}
	patches.(*gomonkey.Patches).Reset()
	patchesMap.Delete(key)
	wait()
	return true
}

// UnpatchAll removes all patches
func UnpatchAll() {
	patchesMap.Range(func(key, value interface{}) bool {
		value.(*gomonkey.Patches).Reset()
		patchesMap.Delete(key)
		return true
	})
	wait()
}

// wait ensures that the patches for darwin/arm64 are applied to prevent test failures and runtime errors, such as invalid memory address or nil pointer dereference
func wait() {
	if runtime.GOOS == "darwin" {
		// use busy wait instead of time.Sleep which will cause SIGBUS error occasionally
		endTime := time.Now().Add(100 * time.Millisecond)
		for time.Now().Before(endTime) {
		}
	}
}
