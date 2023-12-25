/*
Copyright 2021 The KodeRover Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package executor

import (
	commonconfig "github.com/koderover/zadig/v2/pkg/config"
	"github.com/koderover/zadig/v2/pkg/microservice/packager/core/service"
	"github.com/koderover/zadig/v2/pkg/setting"
	"github.com/koderover/zadig/v2/pkg/tool/log"
)

func Execute() error {
	log.Init(&log.Config{
		Level:       commonconfig.LogLevel(),
		NoCaller:    true,
		NoLogLevel:  true,
		Development: commonconfig.Mode() != setting.ReleaseMode,
	})

	packager, err := service.NewPackager()
	if err != nil {
		log.Errorf("Failed to start packager, error: %s", err)
		return err
	}

	if err := packager.Exec(); err != nil {
		log.Errorf("Failed to run packager, error: %s", err)
		return err
	}

	return nil
}
