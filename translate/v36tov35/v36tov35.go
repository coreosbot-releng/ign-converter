// Copyright 2020 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package v36tov35

import (
	"fmt"
	"reflect"

	"github.com/coreos/ignition/config/util"
	"github.com/coreos/ignition/v2/config/translate"
	"github.com/coreos/ignition/v2/config/v3_5/types"

	old_types "github.com/coreos/ignition/v2/config/v3_6/types"
	"github.com/coreos/ignition/v2/config/validate"
)

// Copy of github.com/coreos/ignition/v2/config/v3_5/translate/translate.go
// with the types & old_types imports reversed (the referenced file translates
// from 3.5 -> 3.6 but as a result only touches fields that are understood by
// the 3.5 spec).
func translateFileEmbedded1(old old_types.FileEmbedded1) (ret types.FileEmbedded1) {
	tr := translate.NewTranslator()
	tr.Translate(&old.Append, &ret.Append)
	tr.Translate(&old.Contents, &ret.Contents)
	if old.Mode != nil {
		// Since fixing #2024 we now have to mask for the stabilized specs
		// to reduce security risks of applying permissions that were not applied
		// before the fix was implemented.
		// We support the special mode bits for specs >=3.6.0, so if
		// the user provides special mode bits in an Ignition config
		// with the version < 3.6.0, then we need to explicitly mask
		// those bits out during translation.
		ret.Mode = util.IntToPtr(*old.Mode & ^07000)
	}
	return
}

func translateDirectoryEmbedded1(old old_types.DirectoryEmbedded1) (ret types.DirectoryEmbedded1) {
	if old.Mode != nil {
		// Since fixing #2024 we now have to mask for the stabilized specs
		// to reduce security risks of applying permissions that were not applied
		// before the fix was implemented.
		// We support the special mode bits for specs >=3.6.0, so if
		// the user provides special mode bits in an Ignition config
		// with the version < 3.6.0, then we need to explicitly mask
		// those bits out during translation.
		ret.Mode = util.IntToPtr(*old.Mode & ^07000)
	}
	return
}
func translateIgnition(old old_types.Ignition) (ret types.Ignition) {
	// use a new translator so we don't recurse infinitely
	translate.NewTranslator().Translate(&old, &ret)
	ret.Version = types.MaxVersion.String()
	return
}

func translateConfig(old old_types.Config) (ret types.Config) {
	tr := translate.NewTranslator()
	tr.AddCustomTranslator(translateIgnition)
	tr.AddCustomTranslator(translateDirectoryEmbedded1)
	tr.AddCustomTranslator(translateFileEmbedded1)
	tr.Translate(&old, &ret)
	return
}

// end copied Ignition v3_6/translate block

// Translate translates Ignition spec config v3.6 to spec v3.5
func Translate(cfg old_types.Config) (types.Config, error) {
	rpt := validate.ValidateWithContext(cfg, nil)
	if rpt.IsFatal() {
		return types.Config{}, fmt.Errorf("invalid input config:\n%s", rpt.String())
	}

	err := checkValue(reflect.ValueOf(cfg))
	if err != nil {
		return types.Config{}, err
	}

	res := translateConfig(cfg)

	// Sanity check the returned config
	oldrpt := validate.ValidateWithContext(res, nil)
	if oldrpt.IsFatal() {
		return types.Config{}, fmt.Errorf("converted spec has unexpected fatal error:\n%s", oldrpt.String())
	}
	return res, nil
}

func checkValue(v reflect.Value) error {
	if !v.IsValid() {
		return nil
	}

	// v3.6 introduced arbitrary custom clevis pin support
	if v.Type() == reflect.TypeOf(old_types.ClevisCustom{}) {
		// Check if Pin field is set
		pinField := v.FieldByName("Pin")
		if pinField.IsValid() && !pinField.IsNil() {
			pinValue := pinField.Elem().String()
			// v3.5 only supports tpm2, tang, and sss
			if pinValue != "tpm2" && pinValue != "tang" && pinValue != "sss" {
				return fmt.Errorf("invalid input config: arbitrary custom clevis pin '%s' is not supported in spec v3.5", pinValue)
			}
		}
	}

	// Recursively check nested structures
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			err := checkValue(v.Field(i))
			if err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			err := checkValue(v.Index(i))
			if err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if !v.IsNil() {
			err := checkValue(v.Elem())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
